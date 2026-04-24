package ui

import "github.com/charmbracelet/lipgloss"

// Phosphor-CRT telemetry palette — shared across the Atlas TUI suite.
var (
	ColBG       = lipgloss.Color("#000000")
	ColChrome   = lipgloss.Color("#3A3226")
	ColDim      = lipgloss.Color("#7A6A4A")
	ColText     = lipgloss.Color("#D9C79C")
	ColAmber    = lipgloss.Color("#FFB000")
	ColAmberHot = lipgloss.Color("#FF7A00")
	ColRed      = lipgloss.Color("#FF3D4A")
	ColCyan     = lipgloss.Color("#7DE3FF")
	ColGreen    = lipgloss.Color("#84F5A3")
	ColPaper    = lipgloss.Color("#F5E6D3")
)

var (
	sBorder       = lipgloss.NewStyle().Foreground(ColChrome)
	sLabel        = lipgloss.NewStyle().Foreground(ColDim)
	sText         = lipgloss.NewStyle().Foreground(ColText)
	sValue        = lipgloss.NewStyle().Foreground(ColCyan).Bold(true)
	sPaper        = lipgloss.NewStyle().Foreground(ColPaper).Bold(true)
	sAmber        = lipgloss.NewStyle().Foreground(ColAmber).Bold(true)
	sHot          = lipgloss.NewStyle().Foreground(ColAmberHot).Bold(true)
	sCrit         = lipgloss.NewStyle().Foreground(ColRed).Bold(true)
	sGood         = lipgloss.NewStyle().Foreground(ColGreen).Bold(true)
	sDim          = lipgloss.NewStyle().Foreground(ColDim)
	sRec          = lipgloss.NewStyle().Foreground(ColRed).Bold(true)
	sSectionTitle = lipgloss.NewStyle().Foreground(ColPaper).Bold(true)
	sSectionKey   = lipgloss.NewStyle().Foreground(ColAmber).Bold(true)

	sFooterKey  = lipgloss.NewStyle().Foreground(ColAmber).Bold(true)
	sFooterText = lipgloss.NewStyle().Foreground(ColDim)

	sMastTitle  = lipgloss.NewStyle().Foreground(ColAmber).Bold(true)
	sMastClock  = lipgloss.NewStyle().Foreground(ColPaper).Bold(true)
	sCursor     = lipgloss.NewStyle().Foreground(ColAmber).Bold(true)
	sBigDigit   = lipgloss.NewStyle().Foreground(ColAmber).Bold(true)
	sPromptMark = lipgloss.NewStyle().Foreground(ColAmber).Bold(true)
	sMs         = lipgloss.NewStyle().Foreground(ColDim)
)

// daynightStyle colors the day/night glyph based on the local hour.
func daynightStyle(hour int) (string, lipgloss.Style) {
	switch {
	case hour >= 6 && hour < 18:
		return "☀", sAmber
	case hour >= 18 && hour < 22, hour >= 4 && hour < 6:
		return "☽", sHot
	default:
		return "☾", sDim
	}
}
