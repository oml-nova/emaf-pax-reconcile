package parser

import (
	"fmt"
	"strconv"
)

// ParseCreditHeader extracts file-level metadata from a CREDIT_FILE_HEADER line.
func ParseCreditHeader(line string, mr *MerchantRecord) {
	mr.RecordType = "CREDIT"
	mr.Version = field(line, 23, 28)
	mr.ChainCode = field(line, 105, 111)
	mr.FileCreationDate = parseCreditHeaderDateTime(line)
	mr.ProcessingDate = field(line, 42, 50)
}

func parseCreditHeaderDateTime(line string) string {
	d := field(line, 28, 36) // YYYYMMDD
	t := field(line, 36, 42) // HHMMSS
	if len(d) < 8 || len(t) < 6 {
		return ""
	}
	year, _ := strconv.Atoi(d[0:4])
	month, _ := strconv.Atoi(d[4:6])
	day, _ := strconv.Atoi(d[6:8])
	hour, _ := strconv.Atoi(t[0:2])
	min, _ := strconv.Atoi(t[2:4])
	sec, _ := strconv.Atoi(t[4:6])
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ", year, month, day, hour, min, sec)
}
