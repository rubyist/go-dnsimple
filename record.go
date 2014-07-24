package dnsimple

import (
	"errors"
	"fmt"
	"net/url"
)

type Record struct {
	Id         int    `json:"id,omitempty"`
	Content    string `json:"content,omitempty"`
	Name       string `json:"name,omitempty"`
	TTL        int    `json:"ttl,omitempty"`
	RecordType string `json:"record_type,omitempty"`
	Priority   int    `json:"prio,omitempty"`
	DomainId   int    `json:"domain_id,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

type recordWrapper struct {
	Record Record `json:"record"`
}

func recordPath(domain interface{}, record *Record) string {
	str := fmt.Sprintf("domains/%s/records", domainIdentifier(domain))
	if record != nil {
		str += fmt.Sprintf("/%d", record.Id)
	}
	return str
}

func (client *DNSimpleClient) Records(domain interface{}, name, recordType string) ([]Record, error) {
	reqStr := recordPath(domain, nil)
	v := url.Values{}

	if name != "" {
		v.Add("name", name)
	}

	if recordType != "" {
		v.Add("type", recordType)
	}

	reqStr += "?" + v.Encode()

	wrappedRecords := []recordWrapper{}

	if err := client.get(reqStr, &wrappedRecords); err != nil {
		return []Record{}, err
	}

	records := []Record{}
	for _, record := range wrappedRecords {
		records = append(records, record.Record)
	}

	return records, nil
}

func (client *DNSimpleClient) CreateRecord(domain interface{}, record Record) (Record, error) {
	// pre-validate the Record?
	wrappedRecord := recordWrapper{Record: record}
	returnedRecord := recordWrapper{}

	status, err := client.post(recordPath(domain, nil), wrappedRecord, &returnedRecord)
	if err != nil {
		return Record{}, err
	}

	if status != 201 && status != 200 {
		return Record{}, errors.New(fmt.Sprintf("The API returned an error: HTTP %v", status))
	}

	return returnedRecord.Record, nil
}

func (client *DNSimpleClient) RetrieveRecord(domain interface{}, id int) (Record, error) {
	returnedRecord := recordWrapper{}

	err := client.get(recordPath(domain, &Record{Id: id}), &returnedRecord)
	if err != nil {
		return Record{}, err
	}

	return returnedRecord.Record, nil
}

func (record *Record) Update(client *DNSimpleClient, recordAttributes Record) (Record, error) {
	// pre-validate the Record?
	// name, content, ttl, prio - only things allowed
	wrappedRecord := recordWrapper{Record: Record{
		Name:     recordAttributes.Name,
		Content:  recordAttributes.Content,
		TTL:      recordAttributes.TTL,
		Priority: recordAttributes.Priority}}

	returnedRecord := recordWrapper{}

	status, err := client.put(recordPath(record.DomainId, record), wrappedRecord, &returnedRecord)
	if err != nil {
		return Record{}, err
	}

	if status == 400 {
		return Record{}, errors.New("Invalid Record")
	}

	return returnedRecord.Record, nil
}

func (record *Record) Delete(client *DNSimpleClient) error {
	_, status, err := client.sendRequest("DELETE", recordPath(record.DomainId, record), nil)
	if err != nil {
		return err
	}

	if status == 200 {
		return nil
	}
	return errors.New("Failed to delete domain")
}

func (record *Record) UpdateIP(client *DNSimpleClient, IP string) error {
	newRecord := Record{Content: IP, Name: record.Name}
	_, err := record.Update(client, newRecord)
	return err
}
