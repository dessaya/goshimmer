package unspentoutputs

import (
	"net/http"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/plugins/webapi/value/utils"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
)

// Handler gets the unspent outputs.
func Handler(c echo.Context) error {
	var request Request
	if err := c.Bind(&request); err != nil {
		log.Info(err.Error())
		return c.JSON(http.StatusBadRequest, Response{Error: err.Error()})
	}

	var unspents []UnspentOutput
	for _, strAddress := range request.Addresses {
		address, err := address.FromBase58(strAddress)
		if err != nil {
			log.Info(err.Error())
			continue
		}

		outputids := make([]OutputID, 0)
		// get outputids by address
		for id, outputObj := range valuetransfers.Tangle.OutputsOnAddress(address) {
			defer outputObj.Release()
			output := outputObj.Unwrap()

			// TODO: get inclusion state
			if output.ConsumerCount() == 0 {
				// iterate balances
				var b []utils.Balance
				for _, balance := range output.Balances() {
					b = append(b, utils.Balance{
						Value: balance.Value(),
						Color: balance.Color().String(),
					})
				}

				outputids = append(outputids, OutputID{
					ID:       id.String(),
					Balances: b,
					InclusionState: utils.InclusionState{
						Confirmed: true,
						Conflict:  false,
						Liked:     true,
					},
				})
			}
		}

		unspents = append(unspents, UnspentOutput{
			Address:   strAddress,
			OutputIDs: outputids,
		})
	}

	return c.JSON(http.StatusOK, Response{UnspentOutputs: unspents})
}

// Request holds the addresses to query.
type Request struct {
	Addresses []string `json:"addresses,omitempty"`
	Error     string   `json:"error,omitempty"`
}

// Response is the HTTP response from retreiving value objects.
type Response struct {
	UnspentOutputs []UnspentOutput `json:"unspent_outputs,omitempty"`
	Error          string          `json:"error,omitempty"`
}

// UnspentOutput holds the address and the corresponding unspent output ids
type UnspentOutput struct {
	Address   string     `json:"address"`
	OutputIDs []OutputID `json:"output_ids"`
}

// OutputID holds the output id and its inclusion state
type OutputID struct {
	ID             string               `json:"id"`
	Balances       []utils.Balance      `json:"balances"`
	InclusionState utils.InclusionState `json:"inclusion_state"`
}
