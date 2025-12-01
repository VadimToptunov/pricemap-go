package utils

import (
	"errors"
	"pricemap-go/models"
)

// CurrencyConverter handles currency conversion
type CurrencyConverter struct {
	rates map[string]float64 // Base currency: USD
}

func NewCurrencyConverter() *CurrencyConverter {
	return &CurrencyConverter{
		rates: getDefaultRates(),
	}
}

// Convert converts price from one currency to another
func (cc *CurrencyConverter) Convert(price float64, from, to string) (float64, error) {
	if from == to {
		return price, nil
	}
	
	fromRate, ok := cc.rates[from]
	if !ok {
		return 0, errors.New("unknown currency: " + from)
	}
	
	toRate, ok := cc.rates[to]
	if !ok {
		return 0, errors.New("unknown currency: " + to)
	}
	
	// Convert to USD first, then to target
	usdPrice := price / fromRate
	return usdPrice * toRate, nil
}

// NormalizeToUSD converts any currency to USD
func (cc *CurrencyConverter) NormalizeToUSD(property *models.Property) error {
	if property.Currency == "USD" {
		return nil
	}
	
	converted, err := cc.Convert(property.Price, property.Currency, "USD")
	if err != nil {
		return err
	}
	
	property.Price = converted
	property.Currency = "USD"
	return nil
}

func getDefaultRates() map[string]float64 {
	// Base: USD = 1.0
	// These are approximate rates - in production, use real-time API
	return map[string]float64{
		"USD": 1.0,
		"EUR": 0.92,
		"GBP": 0.79,
		"RUB": 92.0,
		"CNY": 7.2,
		"JPY": 150.0,
		"AUD": 1.52,
		"CAD": 1.35,
		"CHF": 0.88,
		"INR": 83.0,
		"BRL": 4.95,
		"MXN": 17.0,
		"ZAR": 18.5,
		"SEK": 10.5,
		"NOK": 10.8,
		"DKK": 6.85,
		"PLN": 4.0,
		"TRY": 30.0,
		"AED": 3.67,
		"SGD": 1.34,
		"HKD": 7.8,
		"KRW": 1330.0,
		"THB": 35.0,
		"IDR": 15700.0,
		"MYR": 4.7,
		"PHP": 56.0,
		"VND": 24500.0,
	}
}

