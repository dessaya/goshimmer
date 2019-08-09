package mana

import (
	"math"
)

// A calculator that can be used to calculate the changes of mana due to erosion or mana generation.
type Calculator struct {
	decayInterval          float64
	decayFactor            float64
	options                *CalculatorOptions
	tokenSupplyScalefactor float64
}

// Creates a new calculator that can be used to calculate the changes of mana due to erosion or mana generation.
func NewCalculator(decayInterval float64, decayRate float64, optionalOptions ...CalculatorOption) *Calculator {
	return &Calculator{
		// store key settings
		decayInterval: decayInterval,

		// derive important factors ...
		// ... make mana reach exactly the token supply as it's max value (n coins => n mana)
		decayFactor:            1 - decayRate,
		tokenSupplyScalefactor: decayRate / (1 - decayRate),

		// configure optional parameters
		options: DEFAULT_OPTIONS.Override(optionalOptions...),
	}
}

// Returns the amount of mana that was generated by holding the given amount of coins for the given time.
func (calculator *Calculator) GenerateMana(coins uint64, heldTime uint64) (result uint64, roundingError float64) {
	// calculate results
	gainedMana := float64(coins) * calculator.options.ManaScaleFactor * (1 - math.Pow(calculator.decayFactor, float64(heldTime)/calculator.decayInterval))

	// assign rounded results & determine roundingErrors
	result = uint64(gainedMana)
	roundingError = gainedMana - float64(result)

	return
}

// Returns the amount of mana that is left after the erosion of the given amount for the given time.
func (calculator *Calculator) ErodeMana(mana uint64, decayTime uint64) (result uint64, roundingError float64) {
	// if no time has passed -> return unchanged values
	if decayTime == 0 {
		result = mana

		return
	}

	// calculate results
	erodedMana := float64(mana) * math.Pow(calculator.decayFactor, float64(decayTime)/calculator.decayInterval)

	// assign rounded results & determine roundingErrors
	result = uint64(erodedMana)
	roundingError = erodedMana - float64(result)

	return
}