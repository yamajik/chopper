package strings

import "github.com/valyala/fasttemplate"

// Format bulabula
func Format(template string, m map[string]interface{}) string {
	return fasttemplate.New(template, "{", "}").ExecuteString(m)
}
