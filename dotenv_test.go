package dotenv

import (
	"testing"
)

func testReadEnvAndCompare(t *testing.T, envFileName string, expectedValues map[string]string) {
	envMap, err := readAndParseConfig(envFileName, DefaultSeparator)
	if err != nil {
		t.Error("Error reading file")
	}

	t.Log(envMap)

	if len(envMap) != len(expectedValues) {
		t.Error("Didn't get the right size map back")
	}

	for key, value := range expectedValues {
		if envMap[key] != value {
			t.Errorf("Read got one of the keys wrong. Expected: %q got %q", value, envMap[key])
		}
	}
}

func TestReadPlainEnv(t *testing.T) {
	envFileName := "fixtures/plain.env"
	expectedValues := map[string]string{
		"OPTION_A": "1",
		"OPTION_B": "2",
		"OPTION_C": "3",
		"OPTION_D": "4",
		"OPTION_E": "5",
		"OPTION_F": "",
		"OPTION_G": "",
		"OPTION_H": "my string",
	}

	testReadEnvAndCompare(t, envFileName, expectedValues)
}

func TestLoadUnquotedEnv(t *testing.T) {
	envFileName := "fixtures/unquoted.env"
	expectedValues := map[string]string{
		"OPTION_A": "some quoted phrase",
		"OPTION_B": "first one with an unquoted phrase",
		"OPTION_C": "then another one with an unquoted phrase",
		"OPTION_D": "then another one with an unquoted phrase special è char",
		"OPTION_E": "then another one quoted phrase",
	}

	testReadEnvAndCompare(t, envFileName, expectedValues)
}

func TestLoadQuotedEnv(t *testing.T) {
	envFileName := "fixtures/quoted.env"
	expectedValues := map[string]string{
		"OPTION_A": "1",
		"OPTION_B": "2",
		"OPTION_C": "",
		"OPTION_D": "\\n",
		"OPTION_E": "1",
		"OPTION_F": "2",
		"OPTION_G": "",
		"OPTION_H": "\n",
		"OPTION_I": "echo 'asd'",
		"OPTION_J": `first line
second line
third line
and so on`,
		"OPTION_K": "Test#123",
		"OPTION_Z": "last value",
	}

	testReadEnvAndCompare(t, envFileName, expectedValues)
}

func TestLoadExportedEnv(t *testing.T) {
	envFileName := "fixtures/exported.env"
	expectedValues := map[string]string{
		"OPTION_A": "2",
		"OPTION_B": "\\n",
	}

	dotenv := Init(envFileName)
	err := dotenv.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	for key, value := range expectedValues {
		if dotenv.Get(key) != value {
			t.Errorf("Expected: %q got %q", value, dotenv.Get(key))
		}
	}
}
