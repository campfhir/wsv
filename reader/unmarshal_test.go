package reader_test

import (
	"strings"
	"testing"
	"time"

	"github.com/campfhir/wsv/internal"
	"github.com/campfhir/wsv/reader"
)

func TestUnmarsall(t *testing.T) {
	lines := []string{
		`"Age"  "Salary"  "Is Admin"`,
		`39     25210.021	x`,
	}
	data := strings.Join(lines, string('\n'))

	type Employee struct {
		Age   int     `wsv:"Age,format:base10"`
		Money float32 `wsv:"Salary"`
		Admin *bool   `wsv:"Is Admin,format:x|"`
	}
	var s []Employee
	err := reader.Unmarshal([]byte(data), &s)
	if err != nil {
		t.Fatal(err)
	}

	if len(s) != 1 {
		t.Error("expect 1 entry in slice but got", len(s))
		return
	}

	if s[0].Admin == nil || !*s[0].Admin {
		t.Error("expected admin to be true")
	}

	if s[0].Age != 39 {
		t.Error("expect age of 39 but got", s[0].Age)
	}

	if s[0].Money != 25210.021 {
		t.Error("expect salary of 25210.021 but got", s[0].Money)
	}
}

func TestUnmarshalSimpleData(t *testing.T) {
	lines := make([]string, 0)
	lines = append(lines, `"Given Name" "Family Name" "Date of Birth" "Favorite Color"  Age  cool "Is an Admin" Salary "Home Owner" Bought`)
	lines = append(lines, `"Jean Smith" "Le Croix" "01 Jan 2023" "Space Purple" 35 true - "71.299" true "2025-08-02"`)
	lines = append(lines, `"Mary Jane" "Vasquez Rojas" "02 Feb 2021" "Midnight Grey" - false true - true "2022-09-18"`)
	data := strings.Join(lines, string('\n'))
	type hidden_data struct {
		Home_Owner  bool       `wsv:"Home Owner,format:true|false"`
		Bought_Home *time.Time `wsv:"Bought,format:2006-01-02"`
	}
	type Person struct {
		Given_name    string     `wsv:"Given Name"`
		Family_name   string     `wsv:"Family Name"`
		Dob           *time.Time `wsv:"Date of Birth,format:'02 Jan 2006'"`
		FavoriteColor *string    `wsv:"Favorite Color"`
		Age           *int       `wsv:"Age"`
		IsCool        bool       `wsv:"cool,format:true|false"`
		IsAdmin       *bool      `wsv:"Is an Admin,format:true|false"`
		Salary        *float32
		hidden_data
	}
	s := make([]Person, 0)
	err := reader.Unmarshal([]byte(data), &s)
	if err != nil {
		t.Fatal(err)
	}

	if len(s) < 2 {
		t.Error("expected 2 items but only got", len(s))
		return
	}

	p1 := s[0]
	p2 := s[1]

	if p1.Given_name != "Jean Smith" {
		t.Errorf("expect the first person to have the first name of 'Jean Smith' but got '%s'", p1.Given_name)
	}

	if p1.Family_name != "Le Croix" {
		t.Errorf("expect the first person to have the first name of 'Le Croix' but got '%s'", p1.Family_name)
	}

	if p1.FavoriteColor == nil || *p1.FavoriteColor != "Space Purple" {
		t.Errorf("expect the first person to have a favorite color of 'Space Purple' but got '%s'", internal.UnwrapStr(p1.FavoriteColor))
	}

	if p1.Dob == nil || p1.Dob.Format(time.DateOnly) != "2023-01-01" {
		t.Errorf("expect the first person to have a date of birth of '2023-01-01' but got '%s'", internal.UnwrapStr(p1.Dob))
	}

	if p1.Age == nil || *p1.Age != 35 {
		t.Errorf("expect the first person to have a age of '35' but got '%s'", internal.UnwrapStr(p1.Age))
	}

	if !p1.IsCool {
		t.Error("expect the first person to be cool")
	}

	if p1.IsAdmin != nil || (p1.IsAdmin != nil && *p1.IsAdmin) {
		t.Errorf("expect the first person to be a <nil> admin but got '%s'", internal.UnwrapStr(p1.IsAdmin))
	}

	if p1.Salary == nil || *p1.Salary != float32(71.299) {
		t.Errorf("expect the first person to have a salary of 71.299 but got '%s'", internal.UnwrapStr(p1.Salary))
	}

	if !p1.Home_Owner {
		t.Error("expect the first person to be a home owner")
	}

	if p1.Bought_Home == nil || p1.Bought_Home.Format(time.DateOnly) != "2025-08-02" {
		t.Errorf("expect the first person to buy a home at 2025-08-02 but got '%s'", internal.UnwrapStr(p1.Bought_Home))
	}

	if p2.Given_name != "Mary Jane" {
		t.Errorf("expect the first person to have the first name of 'Mary Jane' but got '%s'", p2.Given_name)
	}

	if p2.Family_name != "Vasquez Rojas" {
		t.Errorf("expect the first person to have the first name of 'Vasquez Rojas' but got '%s'", p2.Family_name)
	}

	if p2.FavoriteColor == nil || *p2.FavoriteColor != "Midnight Grey" {
		t.Errorf("expect the first person to have a favorite color of 'Midnight Grey' but got '%s'", internal.UnwrapStr(p2.FavoriteColor))
	}

	if p2.Dob == nil || p2.Dob.Format(time.DateOnly) != "2021-02-02" {
		t.Errorf("expect the first person to have a date of birth of '2023-02-02' but got '%s'", internal.UnwrapStr(p2.Dob))
	}

	if p2.Age != nil {
		t.Errorf("expect the second person to have a age of '<nil>' but got '%s'", internal.UnwrapStr(p2.Age))
	}

	if p2.IsCool {
		t.Error("expect the second person to not be cool")
	}

	if p2.IsAdmin == nil || !*p2.IsAdmin {
		t.Errorf("expect the second person to be an admin but got '%s'", internal.UnwrapStr(p2.IsAdmin))
	}

	if !p2.Home_Owner {
		t.Error("expect the second person to be a home owner")
	}

	if p2.Bought_Home == nil || p2.Bought_Home.Format(time.DateOnly) != "2022-09-18" {
		t.Errorf("expect the first person to buy a home at 2022-09-18 but got '%s'", internal.UnwrapStr(p2.Bought_Home))
	}
}
