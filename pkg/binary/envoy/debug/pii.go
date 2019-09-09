package debug

import (
	"crypto/sha256"
	"fmt"

	"bitbucket.org/creachadair/shell"
)

// process the array of logs by filtering it using format and containsPII, hash the resulting array of logs and return the final array.
// an empty array of strings and an error instance will be returned in the event of an error
func process(logs []string, format string, containsPII map[string]bool) ([]string, error) {
	// filter the valid logs according to the format str
	fieldNames, ok := shell.Split(format)
	if !ok {
		return []string{}, fmt.Errorf("error in splitting format string: %s", format)
	}

	out := []string{}
	for _, log := range logs {
		fieldValues, ok := shell.Split(log)
		if !ok {
			return []string{}, fmt.Errorf("error in splitting log: %s", log)
		}
		// check for correct number of fields, filter and hash the relevent PII fields
		// TODO: add additional check on the field values, i.e, regex check
		if len(fieldValues) == len(fieldNames) {
			requiredvalues := []string{}
			// pick the PII fields and hash the fields
			for j, name := range fieldNames {
				if containsPII[name] {
					fmt.Println("-----inside containsPII loop-----")
					fmt.Println("original value:", fieldValues[j])
					fmt.Println("hashed value: ", hash(fieldValues[j]))
					requiredvalues = append(requiredvalues, hash(fieldValues[j]))
				}
			}
			fmt.Println(shell.Join(requiredvalues))
			out = append(out, shell.Join(requiredvalues))
		}
	}
	return out, nil
}

// TODO: salting the hash
func hash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return string(h.Sum(nil))
}
