package models

type TableData struct {
	Headers   []string                 `json:"headers"`
	Rows      []map[string]interface{} `json:"rows"`
	SheetName string                   `json:"sheet_name"`
}
