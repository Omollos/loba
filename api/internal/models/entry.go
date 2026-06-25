package models

import "time"

// Entry represents a single word, phrase, proverb, or sentence
// in the Loba corpus.
type Entry struct {
	ID                 string     `json:"id"`
	LanguageID         string     `json:"language_id"`
	SourceText         string     `json:"source_text"`
	English            string     `json:"english"`
	PartOfSpeech       *string     `json:"part_of_speech,omitempty"`
	DefinitionSource   *string     `json:"definition_source,omitempty"`
	DefinitionEnglish  *string     `json:"definition_english,omitempty"`
	Explanation        *string     `json:"explanation,omitempty"`
	EnglishEquivalent  *string     `json:"english_equivalent,omitempty"`
	ExampleSource      *string     `json:"example_source,omitempty"`
	ExampleEnglish     *string     `json:"example_english,omitempty"`
	Notes              *string     `json:"notes,omitempty"`
	Category           *string     `json:"category,omitempty"`
	Dialect            *string     `json:"dialect,omitempty"`
	ContributorID      *string    `json:"contributor_id,omitempty"`
	Source             *string     `json:"source,omitempty"`
	SourceURL          *string     `json:"source_url,omitempty"`
	Status             string     `json:"status"`
	VoteScore          int        `json:"vote_score"`
	Rating             string     `json:"rating"`
	FlagReason         *string    `json:"flag_reason,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	ReviewedAt         *time.Time `json:"reviewed_at,omitempty"`
}

// NewEntryRequest represents what arrives when a contributor submits
type NewEntryRequest struct {
	LanguageID        string `json:"language_id"`
	SourceText        string `json:"source_text"`
	English           string `json:"english"`
	PartOfSpeech      string `json:"part_of_speech"`
	DefinitionSource  string `json:"definition_source"`
	DefinitionEnglish string `json:"definition_english"`
	Explanation       string `json:"explanation"`
	EnglishEquivalent string `json:"english_equivalent"`
	ExampleSource     string `json:"example_source"`
	ExampleEnglish    string `json:"example_english"`
	Notes             string `json:"notes"`
	Category          string `json:"category"`
	Dialect           string `json:"dialect"`
	ContributorID     string `json:"contributor_id"`
	Source            string `json:"source"`
	SourceURL         string `json:"source_url"`
}
