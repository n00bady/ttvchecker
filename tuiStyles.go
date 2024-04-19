package main

import "github.com/charmbracelet/lipgloss"

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("55"))

var statusStyle = lipgloss.NewStyle().
	Width(78).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("54")).
	Padding(0, 1, 0, 1).
	Foreground(lipgloss.Color("242"))

var helpStyle = lipgloss.NewStyle().
	Padding(0, 1, 0, 1).
	Width(78).
	Align(lipgloss.Left)
