package config

type Environment string

func (e Environment) IsProduction() bool {
	return e == "production"
}
