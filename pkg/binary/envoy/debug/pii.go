package debug

import (
	"crypto/sha256"
	"fmt"

	"bitbucket.org/creachadair/shell"
)

// process the array of logs by filtering it using format and containsPII, hash the resulting array of logs and return the final array.
// an empty array of strings and an error instance will be returned in the event of an error
func processLogs(logs []string, format string, containsPII map[string]bool) ([]string, error) {
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
		// check for correct number of fields, filter and hash the relevant PII fields
		// TODO: add additional check on the field values, i.e, regex check
		if len(fieldValues) == len(fieldNames) {
			requiredvalues := []string{}
			// pick the PII fields and hash the fields
			for j, name := range fieldNames {
				if containsPII[name] {
					hash, err := hash(fieldValues[j])
					if err != nil {
						return []string{}, fmt.Errorf("error in hashing the field: %s", fieldValues[j])
					}
					requiredvalues = append(requiredvalues, (hash))
				}
			}
			fmt.Println(shell.Join(requiredvalues))
			out = append(out, shell.Join(requiredvalues))
		}
	}
	return out, nil
}

// TODO: salting the hash
func hash(s string) (string, error) {
	h := sha256.New()
	_, err := h.Write([]byte(s))
	if err != nil {
		return "", err
	}
	return string(h.Sum(nil)), nil
}
