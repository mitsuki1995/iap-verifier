package ios

const (
	StringBoolTrue        StringBool  = "true"
	StringBoolFalse       StringBool  = "false"
	IntBoolTrue           IntBool     = "1"
	IntBoolFalse          IntBool     = "0"
	EnvironmentSandbox    Environment = "Sandbox"
	EnvironmentProduction Environment = "Production"
)

type StringBool string
type IntBool string
type Environment string

func (s StringBool) Bool() bool {
	return s == StringBoolTrue
}

func (i IntBool) Bool() bool {
	return i == IntBoolTrue
}

func (e Environment) IsSandbox() bool {
	return e == EnvironmentSandbox
}
