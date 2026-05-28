package hrdb

import (
	"context"
	"errors"
	"strings"
)

// Person represents a record in the corporate HR database.
type Person struct {
	Name           string
	SystemID       string // e.g. yklee
	EmployeeID     string // e.g. 123456
	DepartmentName string
}

var ErrPersonNotFound = errors.New("person not found in HR database")

// Client defines the interface for HR database lookups.
type Client interface {
	Lookup(ctx context.Context, systemID, employeeID, name string) (*Person, error)
}

// MockClient is an in-memory implementation for PoC.
type MockClient struct {
	data []Person
}

func NewMockClient() *MockClient {
	return &MockClient{
		data: []Person{
			{Name: "YK Lee", SystemID: "yklee", EmployeeID: "1001", DepartmentName: "Engineering"},
			{Name: "Alex Kim", SystemID: "akim", EmployeeID: "1002", DepartmentName: "Product"},
			{Name: "Sam Jones", SystemID: "sjones", EmployeeID: "1003", DepartmentName: "Infrastructure"},
		},
	}
}

func (m *MockClient) Lookup(ctx context.Context, systemID, employeeID, name string) (string, string, string, error) {
	for _, p := range m.data {
		if strings.EqualFold(p.SystemID, systemID) &&
			p.EmployeeID == employeeID &&
			strings.EqualFold(p.Name, name) {
			// Construct mock email
			email := strings.ToLower(p.SystemID) + "@example.com"
			return email, p.SystemID, p.DepartmentName, nil
		}
	}
	return "", "", "", ErrPersonNotFound
}
