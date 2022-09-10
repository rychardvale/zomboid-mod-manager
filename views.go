package main

import "fmt"

// view handler for choice Lista
func ListView(m model) string {
	// header
	s := fmt.Sprintf("\n\nInstalled mods: %d\n\n", len(m.Mods))

	for i := m.ListStartIdx; i <= m.ListStartIdx+m.ListRange; i++ {
		if i >= len(m.Mods) {
			break
		}

		cursor := " " // no cursor
		mod := m.Mods[i]
		if m.Cursor == i {
			cursor = cursorStyle.Render(">") // cursor!
			mod = selectedStyle.Render(mod)
		} else {
			mod = unselectedStyle.Render(mod)
		}

		// row
		s += fmt.Sprintf("%s  %s\n", cursor, mod)
	}

	// footer
	s += helpStyle.Render("\n[q] sair.\t[a] adicionar mod.\t[d] deletar mod")

	return s
}

// exiting view
func QuitView(m model) string {
	return fmt.Sprintf(
		"%s",
		cursorStyle.Width(255).Render("So long sucker"),
	)
}

// view for addmod choice
func AddModView(m model) string {
	s := fmt.Sprintf(
		"%s: %s",
		inputStyle.Render("Name"),
		m.NewMod[ModName].View(),
	)

	return s
}
