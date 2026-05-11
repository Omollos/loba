-- ============================================================
-- LOBA CORPUS — DATABASE SCHEMA
-- Version: 1.0
-- Description: Core tables for the Loba open language corpus
-- ============================================================


-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";


-- ============================================================
-- LANGUAGES
-- The anchor table. Every language-specific row in the entire
-- database references this table. Adding a new language is
-- one INSERT here — nothing else changes.
-- ============================================================
CREATE TABLE IF NOT EXISTS languages (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(10) UNIQUE NOT NULL, -- ISO 639-3 e.g. 'luo', 'kik'
    name        TEXT        NOT NULL,        -- e.g. 'Dholuo', 'Kikuyu'
    script      TEXT        NOT NULL,        -- e.g. 'Latin'
    region      TEXT,                        -- e.g. 'Kenya, Tanzania'
    added_at    TIMESTAMPTZ DEFAULT now()
);


-- ============================================================
-- CONTRIBUTORS
-- Everyone who submits or reviews entries.
-- languages[] tracks which languages they speak —
-- useful later for routing entries to bilingual reviewers.
-- ============================================================
CREATE TABLE IF NOT EXISTS contributors (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(50) UNIQUE NOT NULL,
    email           TEXT        UNIQUE,
    github_handle   VARCHAR(50),
    languages       UUID[],     -- array of language IDs they speak
    joined_at       TIMESTAMPTZ DEFAULT now()
);


-- ============================================================
-- ENTRIES
-- The core corpus table. Every word, phrase, proverb,
-- or sentence submitted to Loba lives here.
-- ============================================================
CREATE TABLE IF NOT EXISTS entries (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Multi-language support
    language_id         UUID        NOT NULL REFERENCES languages(id),

    -- Core content
    source_text         TEXT        NOT NULL, -- the word/phrase in source language
    english             TEXT        NOT NULL, -- english translation
    part_of_speech      VARCHAR(30),          -- noun, verb, phrase, proverb etc

    -- Enriched content
    explanation         TEXT,                 -- deeper meaning / cultural context
    english_equivalent  TEXT,                 -- e.g. 'Need has no boundary'
    example_source      TEXT,                 -- example sentence in source language
    example_english     TEXT,                 -- example sentence translation
    notes               TEXT,                 -- pronunciation, tone, register

    -- Classification
    category            VARCHAR(50),          -- Greetings, Proverbs, Family etc
    dialect             VARCHAR(50),          -- Kisumu, Siaya, Homa Bay etc

    -- Attribution and sourcing
    contributor_id      UUID        REFERENCES contributors(id),
    source              TEXT,                 -- e.g. 'Luo Sayings, Lake Publishers'
    source_url          TEXT,                 -- e.g. 'archive.org/details/luo-sayings'

    -- Moderation
    status              VARCHAR(20) DEFAULT 'pending'
                        CHECK (status IN ('pending', 'approved', 'flagged')),
    vote_score          INT         DEFAULT 0,

    -- Timestamps
    created_at          TIMESTAMPTZ DEFAULT now(),
    reviewed_at         TIMESTAMPTZ
);


-- ============================================================
-- VOTES
-- Crowd-sourced accuracy checking.
-- One vote per contributor per entry — enforced by PRIMARY KEY.
-- Vote score on entries table is updated separately.
-- ============================================================
CREATE TABLE IF NOT EXISTS votes (
    entry_id    UUID        NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    voter_id    UUID        NOT NULL REFERENCES contributors(id),
    vote        SMALLINT    NOT NULL CHECK (vote IN (-1, 1)),
    voted_at    TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (entry_id, voter_id)
);


-- ============================================================
-- INDEXES
-- Added after tables for performance on common queries.
-- ============================================================

-- Full text search on English translations
CREATE INDEX IF NOT EXISTS idx_entries_english_fts
    ON entries USING gin(to_tsvector('english', english));

-- Full text search on source language text
CREATE INDEX IF NOT EXISTS idx_entries_source_fts
    ON entries USING gin(to_tsvector('simple', source_text));

-- Filtered listing — the most common query pattern
CREATE INDEX IF NOT EXISTS idx_entries_filter
    ON entries (language_id, status, category, created_at DESC);

-- Contributor leaderboard queries
CREATE INDEX IF NOT EXISTS idx_entries_contributor
    ON entries (contributor_id);

-- Vote lookups
CREATE INDEX IF NOT EXISTS idx_votes_entry
    ON votes (entry_id);


-- ============================================================
-- SEED DATA
-- The very first row — Dholuo as the founding language.
-- ============================================================
INSERT INTO languages (code, name, script, region)
VALUES ('luo', 'Dholuo', 'Latin', 'Kenya, Tanzania')
ON CONFLICT (code) DO NOTHING;