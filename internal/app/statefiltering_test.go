package app_test

import (
	"fmt"
	"github.com/hedhyw/json-log-viewer/internal/pkg/source"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hedhyw/json-log-viewer/assets"
	"github.com/hedhyw/json-log-viewer/internal/app"
	"github.com/hedhyw/json-log-viewer/internal/pkg/events"
)

func TestStateFiltering(t *testing.T) {
	t.Parallel()

	model, source := newTestModel(t, assets.ExampleJSONLog())
	defer source.Close()

	model = handleUpdate(model, tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'f'},
	})
	_, ok := model.(app.StateFilteringModel)
	assert.Truef(t, ok, "%s", model)

	t.Run("input_hotkeys", func(t *testing.T) {
		t.Parallel()

		model := handleUpdate(model, tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'q'},
		})

		model = handleUpdate(model, tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'f'},
		})

		_, ok := model.(app.StateFilteringModel)
		assert.Truef(t, ok, "%s", model)
	})

	t.Run("returned", func(t *testing.T) {
		t.Parallel()

		model := handleUpdate(model, tea.KeyMsg{
			Type: tea.KeyEsc,
		})

		_, ok := model.(app.StateLoadedModel)
		assert.Truef(t, ok, "%s", model)
	})

	t.Run("empty_input", func(t *testing.T) {
		t.Parallel()

		model := handleUpdate(model, tea.KeyMsg{
			Type: tea.KeyEnter,
		})

		_, ok := model.(app.StateLoadedModel)
		require.Truef(t, ok, "%s", model)
	})

	t.Run("stringer", func(t *testing.T) {
		t.Parallel()

		stringer, ok := model.(fmt.Stringer)
		if assert.True(t, ok) {
			assert.Contains(t, stringer.String(), "StateFiltering")
		}
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		model := handleUpdate(model, events.ErrorOccuredMsg{Err: getTestError()})

		_, ok := model.(app.StateErrorModel)
		assert.Truef(t, ok, "%s", model)
	})

	t.Run("navigation", func(t *testing.T) {
		t.Parallel()

		model := handleUpdate(model, tea.KeyMsg{
			Type: tea.KeyUp,
		})

		_, ok := model.(app.StateFilteringModel)
		assert.Truef(t, ok, "%s", model)
	})
}

func TestStateFilteringReset(t *testing.T) {
	t.Parallel()

	const termIncluded = "included"

	const jsonFile = `
	{"time":"1970-01-01T00:00:00.00","level":"INFO","message": "` + termIncluded + `"}
	`

	setup := func() (tea.Model, *source.Source) {
		model, source := newTestModel(t, []byte(jsonFile))

		rendered := model.View()
		assert.Contains(t, rendered, termIncluded)

		// Open filter.
		model = handleUpdate(model, tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'f'},
		})

		_, ok := model.(app.StateFilteringModel)
		assert.Truef(t, ok, "%s", model)

		// Filter to exclude everything.
		model = handleUpdate(model, tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune(termIncluded + "_not_found"),
		})
		model = handleUpdate(model, tea.KeyMsg{
			Type: tea.KeyEnter,
		})

		_, ok = model.(app.StateFilteredModel)
		assert.Truef(t, ok, "%s", model)
		return model, source
	}

	t.Run("record_not_included", func(t *testing.T) {
		t.Parallel()

		model, source := setup()
		defer source.Close()

		rendered := model.View()

		index := strings.Index(rendered, "filtered 0 by:")
		if assert.Positive(t, index) {
			rendered = rendered[:index]
		}

		assert.NotContains(t, rendered, termIncluded)

		// Come back
		model = handleUpdate(model, tea.KeyMsg{
			Type: tea.KeyEsc,
		})

		_, ok := model.(app.StateLoadedModel)
		assert.Truef(t, ok, "%s", model)

		// Assert.
		rendered = model.View()
		assert.Contains(t, rendered, termIncluded)
	})

	t.Run("record_not_included", func(t *testing.T) {
		t.Parallel()

		model, source := setup()
		defer source.Close()

		// Try to open a record where there are no records.
		model = handleUpdate(model, tea.KeyMsg{
			Type: tea.KeyEnter,
		})

		assert.NotNil(t, model)

		_, ok := model.(app.StateLoadedModel)
		assert.Truef(t, ok, "%s", model)
	})
}
