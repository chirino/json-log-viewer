package app

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hedhyw/json-log-viewer/internal/pkg/config"
	"slices"
)

// rowGetter renders the row.
type rowGetter interface {
	// Row return a rendered table row.
	Row(cfg *config.Config, i int) table.Row
	Len() int
}

// lazyTableModel lazily renders table rows.
type lazyTableModel struct {
	*Application

	table table.Model

	minRenderedRows int
	entries         rowGetter
	lastCursor      int
	offset          int
	reverse         bool
	follow          bool

	renderedRows []table.Row
}

type EntriesUpdateMsg struct {
	Entries rowGetter
}

// View implements tea.Model.
func (m lazyTableModel) View() string {
	return m.table.View()
}

// Update implements tea.Model.
func (m lazyTableModel) Update(msg tea.Msg) (lazyTableModel, tea.Cmd) {
	var cmd tea.Cmd

	render := false
	switch msg := msg.(type) {
	case tea.KeyMsg:

		if key.Matches(msg, defaultKeys.Reverse) {
			m.reverse = !m.reverse
			render = true
		}

		increaseOffset := func() {
			maxO := max(m.entries.Len()-m.table.Height(), 0)
			o := min(m.offset+1, maxO)
			if o != m.offset {
				m.offset = o
				render = true
			} else {
				m.follow = true
			}
		}
		decreaseOffset := func() {
			o := max(m.offset-1, 0)
			if o != m.offset {
				m.offset = o
				render = true
			}
		}
		if m.reverse {
			x := increaseOffset
			increaseOffset = decreaseOffset
			decreaseOffset = x
		}
		if key.Matches(msg, defaultKeys.Down) {
			m.follow = false
			if m.table.Cursor()+1 == m.table.Height() {
				increaseOffset()
			}
		}
		if key.Matches(msg, defaultKeys.Up) {
			m.follow = false
			if m.table.Cursor() == 0 {
				decreaseOffset()
			}
		}

	case EntriesUpdateMsg:
		m.entries = msg.Entries
		render = true
	}
	m.table, cmd = m.table.Update(msg)

	if m.table.Cursor() != m.lastCursor {
		render = true
	}

	if render {
		m = m.RenderedRows()
	}

	return m, cmd
}

func (m lazyTableModel) ViewPortCursor() int {
	if m.reverse {
		viewSize := m.ViewPortEnd() - m.ViewPortStart()
		return m.offset + (viewSize - 1 - m.table.Cursor())
	} else {
		return m.offset + m.table.Cursor()
	}
}

func (m lazyTableModel) ViewPortStart() int {
	return m.offset
}

func (m lazyTableModel) ViewPortEnd() int {
	return min(m.offset+m.table.Height(), m.entries.Len())
}

func (m lazyTableModel) RenderedRows() lazyTableModel {
	if m.follow {
		m.offset = max(0, m.entries.Len()-m.table.Height())
	}
	end := min(m.offset+m.table.Height(), m.entries.Len())

	m.renderedRows = []table.Row{}
	for i := m.offset; i < end; i++ {
		m.renderedRows = append(m.renderedRows, m.entries.Row(m.Config, i))
	}

	if m.reverse {
		slices.Reverse(m.renderedRows)
	}

	m.table.SetRows(m.renderedRows)
	if m.follow {
		if m.reverse {
			m.table.GotoTop()
		} else {
			m.table.GotoBottom()
		}
	}

	m.lastCursor = m.table.Cursor()

	return m
}
