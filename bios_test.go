/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * bios_test.go
 */

package hatchet

import (
	"testing"
)

func TestGenerateBio(t *testing.T) {
	bio := generateBio()

	// Check that all fields are non-empty
	if bio.FirstName == "" {
		t.Error("FirstName is empty")
	}
	if bio.Emails[0] == "" {
		t.Error("EmailAddress is empty")
	}
	if bio.Phones[0] == "" {
		t.Error("PhoneNumber is empty")
	}
	if bio.SSN == "" {
		t.Error("SSN is empty")
	}
	if bio.CreditCards[0] == "" {
		t.Error("CreditCardNumber is empty")
	}
	if bio.State == "" {
		t.Error("StateName is empty")
	}
	if bio.URL == "" {
		t.Error("PersonalWebsite is empty")
	}
	if bio.Intro == "" {
		t.Error("ShortDescription is empty")
	}
	if bio.Age < 1 || bio.Age > 100 {
		t.Error("Age is out of range")
	}
	if bio.Intro == "" {
		t.Error("ShortDescription is empty")
	}

	t.Log(bio.Intro)
}
