package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/enescakir/emoji"
	"github.com/rivo/uniseg"
)

type model struct {
	tokens        int
	dollars       int
	ticksLeft     int
	spinnerMsg    string
	spinning      bool
	isDay         bool
	windowWidth   int
	windowHeight  int
	jackpotAmount int
	slot          Slot
	isFreeSpin    bool
	symbolDT      SymbolDropTable
	multiplier    int
}

type SymbolDropTable map[emoji.Emoji]int

type TickMsg time.Time
type Symbol emoji.Emoji
type Slot [3]Symbol

func initTable() SymbolDropTable {
	return map[emoji.Emoji]int{
		emoji.Keycap1:       100,
		emoji.Keycap2:       100,
		emoji.Keycap3:       100,
		emoji.Keycap4:       100,
		emoji.Keycap5:       100,
		emoji.Keycap6:       0,
		emoji.Keycap7:       0,
		emoji.Keycap8:       0,
		emoji.Keycap9:       0,
		emoji.Keycap10:      0,
		emoji.HoneyPot:      50,
		emoji.Joker:         50,
		emoji.HundredPoints: 50,
		emoji.GameDie:       50,
		emoji.FreeButton:    50,
		emoji.Skull:         50,
		emoji.Sunrise:       50,
		emoji.FullMoon:      50,
		emoji.Fire:          50,
		emoji.Bomb:          50,
	}
}

func (s SymbolDropTable) rollTable() Symbol {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	roll := r.Float64() // generate a random float between [0 and 1)

	var cumulativeProbability float64
	for e, prob := range s {
		cumulativeProbability += float64(prob)
		if roll < cumulativeProbability {
			return Symbol(e)
		}
	}

	// This should never happen if probabilities sum to 1, but just in case:
	return Symbol(emoji.WhiteFlag)
}

func (s Symbol) toEmoji() emoji.Emoji {
	return emoji.Emoji(s)
}

func (s Symbol) toString() string {
	emojiString := emoji.Emoji(s).String()
	emojiWidth := uniseg.StringWidth(emojiString)
	maxWidth := 5

	// need to do string magic here because every emoji is not the same width
	if emojiWidth < maxWidth {
		return emojiString + strings.Repeat(" ", maxWidth-emojiWidth)
	}

	return emojiString
}

func (s Symbol) toInt() int {
	switch s.toEmoji() {
	case emoji.Keycap1:
		return 1
	case emoji.Keycap2:
		return 2
	case emoji.Keycap3:
		return 3
	case emoji.Keycap4:
		return 4
	case emoji.Keycap5:
		return 5
	case emoji.Keycap6:
		return 6
	case emoji.Keycap7:
		return 7
	case emoji.Keycap8:
		return 8
	case emoji.Keycap9:
		return 9
	case emoji.Keycap10:
		return 10
	}
	return -1
}

func (s Slot) getFirstNumber() Symbol {
	if s[0] != Symbol(emoji.Joker) {
		return s[0]
	}
	if s[1] != Symbol(emoji.Joker) {
		return s[1]
	}
	if s[2] != Symbol(emoji.Joker) {
		return s[2]
	}
	return Symbol(emoji.Joker)
}

func (s Slot) allSymbolsMatch() bool {
	if s[0] == s[1] && s[0] == s[2] {
		return true
	}
	return false
}

// either returns the number they match if they all match or -1 if they don't match
func (s Slot) allNumbersMatch() int {
	// handle joker logic
	if (s[0] == Symbol(emoji.Joker) && s[1] == s[2]) ||
		(s[1] == Symbol(emoji.Joker) && s[0] == s[2]) ||
		(s[2] == Symbol(emoji.Joker) && s[0] == s[1]) {
		return s.getFirstNumber().toInt()
	}

	if s[0] == Symbol(emoji.Joker) && s[1] == Symbol(emoji.Joker) ||
		s[1] == Symbol(emoji.Joker) && s[2] == Symbol(emoji.Joker) ||
		s[0] == Symbol(emoji.Joker) && s[2] == Symbol(emoji.Joker) {
		return s.getFirstNumber().toInt()
	}

	// no jokers here :)
	if s[0] == s[1] && s[0] == s[2] {
		return s[0].toInt()
	}

	// no match
	return -1
}

func (s Slot) toStringArray() []string {
	strs := []string{}
	for i := 0; i < 3; i++ {
		strs = append(strs, s[i].toString())
	}
	return strs
}

func initSlot() Slot {
	slot := Slot{}
	for i := 0; i < 3; i++ {
		slot[i] = Symbol(emoji.SlotMachine)
	}
	return slot
}

func (m *model) SpinSlots() {
	if m.ticksLeft > 60 {
		m.slot[0] = m.symbolDT.rollTable()
	}
	if m.ticksLeft > 30 {
		m.slot[1] = m.symbolDT.rollTable()

	}
	m.slot[2] = m.symbolDT.rollTable()
}

func (m *model) handleWin() {
	switch m.slot[0] {
	case Symbol(emoji.HoneyPot):
		m.dollars += m.jackpotAmount
		m.spinnerMsg = "🎉🎉🎉YOU HIT THE JACKPOT!!! YOU WON " + strconv.Itoa(
			m.jackpotAmount,
		) + " dollars!!!!🎉🎉🎉"
		m.jackpotAmount = 0
	case Symbol(emoji.HundredPoints):
		m.dollars += 100
		m.spinnerMsg = "🎉 Nice 100! You won 100 dollars!"
	case Symbol(emoji.Joker):
		m.spinnerMsg = "Hahahahahahahaha " + emoji.RollingOnTheFloorLaughing.String() + emoji.RollingOnTheFloorLaughing.String() + emoji.RollingOnTheFloorLaughing.String()
	case Symbol(emoji.GameDie):
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		roll := r.Intn(6) + 1 // generate a random int between [1 and 6]
		m.multiplier = roll
		m.spinnerMsg = "Your new multiplier is x" + strconv.Itoa(roll)
	case Symbol(emoji.FreeButton):
		m.isFreeSpin = true
		m.spinnerMsg = "Nice! You got a free spin! Go ahead and reroll!"
	case Symbol(emoji.Keycap1):
		fallthrough
	case Symbol(emoji.Keycap2):
		fallthrough
	case Symbol(emoji.Keycap3):
		fallthrough
	case Symbol(emoji.Keycap4):
		fallthrough
	case Symbol(emoji.Keycap5):
		fallthrough
	case Symbol(emoji.Keycap6):
		fallthrough
	case Symbol(emoji.Keycap7):
		fallthrough
	case Symbol(emoji.Keycap8):
		fallthrough
	case Symbol(emoji.Keycap9):
		fallthrough
	case Symbol(emoji.Keycap10):
		wonDollars := m.slot.getFirstNumber().toInt() * m.multiplier
		m.dollars += wonDollars
		m.spinnerMsg = "🎉You won " + strconv.Itoa(wonDollars) + " dollars!🎉"
	}
}

func (m *model) finishSpin() {
	m.spinning = false
	if m.slot.allSymbolsMatch() || m.slot.allNumbersMatch() != -1 {
		m.handleWin()
	} else {
		m.spinnerMsg = "Try again :("
		m.jackpotAmount += 10
	}

}

func (m *model) doTick() tea.Cmd {
	m.SpinSlots()
	m.ticksLeft--
	if m.ticksLeft == 0 {
		m.finishSpin()
		return nil
	}
	return tea.Tick(time.Second/40, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func initialModel() model {
	// TODO get the initial model from local storage
	return model{
		tokens:        100,
		dollars:       0,
		ticksLeft:     0,
		spinning:      false,
		spinnerMsg:    "",
		isDay:         true,
		jackpotAmount: 0,
		multiplier:    1,
		isFreeSpin:    false,
		slot:          initSlot(),
		symbolDT:      initTable(),
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

	case TickMsg:
		return m, m.doTick()

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case " ":
			if m.spinning {
				m.spinnerMsg = "Chill, the spinner is still spinning!"
				return m, nil
			} else if m.tokens == 0 && !m.isFreeSpin {
				m.spinnerMsg = "You have no more tokens!"
				return m, nil
			} else if m.tokens < 0 {
				m.spinnerMsg = "How the heck do you have negative tokens?"
				return m, nil
			}

			// free spin logic
			if m.isFreeSpin {
				m.spinnerMsg = "Enjoy the free spin!"
				m.isFreeSpin = false
			} else {
				m.spinnerMsg = ""
				m.tokens--
			}

			m.ticksLeft = 90
			m.spinning = true
			return m, m.doTick()
		}
	}

	return m, nil
}

func (m model) View() string {

	str := "         " + lipgloss.NewStyle().
		Render(m.slot.toStringArray()...) +
		"  x" + strconv.Itoa(m.multiplier) + " mult"

	tokenStr := strconv.Itoa(m.tokens)
	dollarsStr := strconv.Itoa(m.dollars)
	jackpotAmountStr := strconv.Itoa(m.jackpotAmount)

	timeStr := "It is currently daytime! You are safe " + emoji.Sunrise.String()
	if !m.isDay {
		timeStr = "It is currently nighttime! Watch out for the Wheel of Misfortune! " + emoji.Skull.String()
	}

	consoleTxt := "\n\n" + m.spinnerMsg +
		"\nPress the spacebar to spin!" +
		"\nYou have " + tokenStr + " tokens left" +
		"\nYou have " + dollarsStr + " dollars" +
		"\n" + timeStr +
		"\nThe current jackpot is worth " + jackpotAmountStr + " dollars!" +
		"\nPress 'q' to quit" + "\n"
	console := lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).Height(10)

	style := lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center)

	screenTxt := lipgloss.JoinVertical(
		lipgloss.Center,
		style.Render(str),
		console.Render(consoleTxt),
	)

	screen := lipgloss.NewStyle().
		Width(m.windowWidth-2).
		Height(m.windowHeight-2).
		Align(lipgloss.Center, lipgloss.Center).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("2")).
		Render(screenTxt)

	return screen
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithFPS(120), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
