package handler

import (
	"strconv"
	"strings"
)

func bindScriptEnvs(envs []string, vars map[string]interface{}, scriptEnvs map[string][]string) []string {
	// vars: variable name -> value
	// envs: "FOO=value"
	// scriptEnvs: variable name -> bound environment variable names
	for k, envNames := range scriptEnvs {
		switch v := vars[k].(type) {
		case string:
			for _, e := range envNames {
				envs = append(envs, e+"="+v)
			}
		case bool:
			a := strconv.FormatBool(v)
			for _, e := range envNames {
				envs = append(envs, e+"="+a)
			}
		case []string:
			a := strings.Join(v, ",")
			for _, e := range envNames {
				envs = append(envs, e+"="+a)
			}
		}
	}
	return envs
}
