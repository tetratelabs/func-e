package debug

import (
	"fmt"
	"bitbucket.org/creachadair/shell"
	"crypto/sha256"
)

// filter the array of logs using formatStr and validFieldNames, hash the resulting array of logs and return the final array
func process(logsArr []string, formatStr string, validFieldNames map[string]bool) (processedLogs []string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("process func panicking, recovering")
		}
	} ()

	fieldNames, ok := shell.Split(formatStr)

	if !ok {
		panic(fmt.Sprint("error in splitting format string: %s", formatStr))
	}

	testForValidLog :=  func(log string) bool {
		fields, ok := shell.Split(log)
		if !ok {
			panic(fmt.Sprint("error in splitting log: %s", log))
		}
		// TODO: consider using regex to check fields instead of merely verifiying the number of fields
		return len(fields) == len(fieldNames)
	}
	validLogs := filter(logsArr, testForValidLog)

	// filter and hash each log in the array of valid logs with relevant fields
	for i,_ := range(validLogs) {
		fields, ok := shell.Split(validLogs[i])
		if !ok {
			panic(fmt.Sprint("error in splitting a valid log: %s", validLogs[i]))
		}

		// filter fields by validFieldNames
		filteredFields := []string{}
		for j,_ := range(fields) {
			if (validFieldNames[fieldNames[j]]) {
				filteredFields = append(filteredFields,fields[j])
			}
		}
		validLog := shell.Join(filteredFields)

		// hash log
		validLog = hash(validLog)
		processedLogs = append(processedLogs, validLog)
	}

	return
}

func filter(strSlice []string, test func(string) bool) (filteredStrSlice []string) {
	for _,s := range strSlice {
		if test(s) {
			filteredStrSlice = append(filteredStrSlice, s)
		}
	}
	return
}

// TODO: salting the hash
func hash(s string) (hashedString string) {
	h := sha256.New()
	h.Write([]byte(s))
	hashedString = string(h.Sum(nil))
	return
}