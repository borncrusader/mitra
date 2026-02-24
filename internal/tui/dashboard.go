package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"mitra/internal/config"
	"mitra/internal/proto"
	"mitra/internal/tmux"
	"mitra/internal/util"
)

type paneView int

const (
	paneRepos paneView = iota
	paneSessions
)

type dataLoadedMsg struct {
	repos     []*proto.Repo
	worktrees []*proto.Worktree
	sessions  []*proto.Session
	err       error
}

type sessionDetachedMsg struct{ err error }

type tickMsg time.Time

// repoItem

type repoItem struct {
	repo      *proto.Repo
	worktrees []*proto.Worktree
}

func (i repoItem) Title() string {
	return fmt.Sprintf("%s/%s", i.repo.Owner, i.repo.Repo)
}

func (i repoItem) Description() string {
	if len(i.worktrees) == 0 {
		return "No worktrees"
	}

	var branches []string
	for _, wt := range i.worktrees {
		branches = append(branches, wt.Branch)
	}

	if len(branches) > 5 {
		return fmt.Sprintf("%d worktrees: %s, ...", len(branches), strings.Join(branches[:5], ", "))
	}
	return fmt.Sprintf("%d worktrees: %s", len(branches), strings.Join(branches, ", "))
}

func (i repoItem) FilterValue() string { return i.Title() }

// sessionItem

type sessionItem struct {
	repoPath string
	branch   string
}

func (i sessionItem) Title() string       { return i.repoPath }
func (i sessionItem) Description() string { return i.branch }
func (i sessionItem) FilterValue() string { return i.repoPath }

// styles

var (
	activeTitleStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("15")).
				Foreground(lipgloss.Color("0")).
				Padding(0, 1)
	inactiveTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Padding(0, 1)
)

// DashboardModel

type DashboardModel struct {
	width             int
	height            int
	quitting          bool
	focusedPane       paneView
	repos             []*proto.Repo
	worktrees         []*proto.Worktree
	sessions          []*proto.Session
	reposList         list.Model
	reposListReady    bool
	sessionsList      list.Model
	sessionsListReady bool
	err               error
	spinner           spinner.Model
	connecting        bool
	nextRetryTime     time.Time
	savedReposIdx     int
	savedSessionsIdx  int
	restoreIdx        bool
}

func NewDashboard() DashboardModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	return DashboardModel{
		connecting: true,
		spinner:    s,
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(loadData, m.spinner.Tick)
}

func tickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func loadData() tea.Msg {
	cfg, err := config.Load()
	if err != nil {
		return dataLoadedMsg{err: fmt.Errorf("failed to load config: %w", err)}
	}

	conn, err := grpc.NewClient("localhost"+cfg.Server.GrpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return dataLoadedMsg{err: fmt.Errorf("failed to connect to server: %w", err)}
	}
	defer util.DeferCheck(conn.Close)

	client := proto.NewMitraServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repoResp, err := client.ListRepos(ctx, &proto.ListReposRequest{})
	if err != nil {
		return dataLoadedMsg{err: fmt.Errorf("failed to list repos: %w", err)}
	}

	worktreeResp, err := client.ListWorktrees(ctx, &proto.ListWorktreesRequest{})
	if err != nil {
		return dataLoadedMsg{err: fmt.Errorf("failed to list worktrees: %w", err)}
	}

	sessionResp, err := client.ListSessions(ctx, &proto.ListSessionsRequest{})
	if err != nil {
		return dataLoadedMsg{err: fmt.Errorf("failed to list sessions: %w", err)}
	}

	return dataLoadedMsg{
		repos:     repoResp.Repos,
		worktrees: worktreeResp.Worktrees,
		sessions:  sessionResp.Sessions,
	}
}

const bannerTopPad = 2

func (m DashboardModel) paneHeight() int {
	// bannerTopPad + 6 banner + 1 empty + 1 footer = 10; split remainder between two panes
	return (m.height - 8 - bannerTopPad) / 2
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h := m.paneHeight()
		if m.reposListReady {
			m.reposList.SetSize(msg.Width-4, h)
		}
		if m.sessionsListReady {
			m.sessionsList.SetSize(msg.Width-4, h)
		}
		return m, nil

	case tea.KeyMsg:
		// Let the focused list consume keys when it's filtering
		if m.focusedPane == paneRepos && m.reposListReady && m.reposList.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.reposList, cmd = m.reposList.Update(msg)
			return m, cmd
		}
		if m.focusedPane == paneSessions && m.sessionsListReady && m.sessionsList.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.sessionsList, cmd = m.sessionsList.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "tab":
			if m.focusedPane == paneRepos {
				m.focusedPane = paneSessions
			} else {
				m.focusedPane = paneRepos
			}
			m.applyFocusStyles()
			return m, nil
		case "enter":
			if m.focusedPane == paneSessions && m.sessionsListReady && len(m.sessions) > 0 {
				idx := m.sessionsList.Index()
				if idx >= 0 && idx < len(m.sessions) {
					sessionID := m.sessions[idx].Id
					m.savedReposIdx = m.reposList.Index()
					m.savedSessionsIdx = idx
					m.restoreIdx = true
					attachCmd := tmux.AttachSessionCmd(sessionID)
					return m, tea.ExecProcess(attachCmd, func(err error) tea.Msg {
						return sessionDetachedMsg{err: err}
					})
				}
			}
		default:
			if m.focusedPane == paneRepos && m.reposListReady {
				var cmd tea.Cmd
				m.reposList, cmd = m.reposList.Update(msg)
				return m, cmd
			}
			if m.focusedPane == paneSessions && m.sessionsListReady {
				var cmd tea.Cmd
				m.sessionsList, cmd = m.sessionsList.Update(msg)
				return m, cmd
			}
		}

	case dataLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.nextRetryTime = time.Now().Add(5 * time.Second)
			return m, tickCmd()
		}
		m.connecting = false
		m.repos = msg.repos
		m.worktrees = msg.worktrees
		m.sessions = msg.sessions

		h := m.paneHeight()
		w := m.width - 4

		// Build repos list
		repoItems := make([]list.Item, len(m.repos))
		for i, repo := range m.repos {
			repoItems[i] = repoItem{
				repo:      repo,
				worktrees: m.getWorktreesForRepo(repo.Id),
			}
		}
		m.reposList = list.New(repoItems, list.NewDefaultDelegate(), w, h)
		m.reposList.Title = "Repositories"
		m.reposList.SetShowStatusBar(true)
		m.reposList.SetFilteringEnabled(true)
		m.reposListReady = true
		if m.restoreIdx {
			m.reposList.Select(m.savedReposIdx)
		}

		// Build sessions list
		worktreeMap := make(map[string]*proto.Worktree)
		for _, wt := range m.worktrees {
			worktreeMap[wt.Id] = wt
		}
		repoMap := make(map[string]*proto.Repo)
		for _, r := range m.repos {
			repoMap[r.Id] = r
		}
		sessionItems := make([]list.Item, len(m.sessions))
		for i, s := range m.sessions {
			repoPath, branch := "unknown", "unknown"
			if wt := worktreeMap[s.WorktreeId]; wt != nil {
				branch = wt.Branch
				if r := repoMap[wt.RepoId]; r != nil {
					repoPath = fmt.Sprintf("%s/%s/%s", r.Host, r.Owner, r.Repo)
				}
			}
			sessionItems[i] = sessionItem{repoPath: repoPath, branch: branch}
		}
		m.sessionsList = list.New(sessionItems, list.NewDefaultDelegate(), w, h)
		m.sessionsList.Title = "Sessions"
		m.sessionsList.SetShowStatusBar(false)
		m.sessionsList.SetFilteringEnabled(true)
		m.sessionsListReady = true
		if m.restoreIdx {
			m.sessionsList.Select(m.savedSessionsIdx)
			m.restoreIdx = false
		}

		m.applyFocusStyles()
		return m, nil

	case sessionDetachedMsg:
		return m, loadData

	case tickMsg:
		if m.connecting {
			if time.Now().After(m.nextRetryTime) {
				return m, loadData
			}
			return m, tickCmd()
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *DashboardModel) applyFocusStyles() {
	if m.reposListReady {
		if m.focusedPane == paneRepos {
			m.reposList.Styles.Title = activeTitleStyle
		} else {
			m.reposList.Styles.Title = inactiveTitleStyle
		}
	}
	if m.sessionsListReady {
		if m.focusedPane == paneSessions {
			m.sessionsList.Styles.Title = activeTitleStyle
		} else {
			m.sessionsList.Styles.Title = inactiveTitleStyle
		}
	}
}

func (m DashboardModel) View() string {
	if m.quitting {
		return ""
	}
	return m.renderDashboard()
}

func (m DashboardModel) renderDashboard() string {
	if m.connecting {
		var lines []string
		if m.nextRetryTime.IsZero() {
			lines = []string{fmt.Sprintf("%s Connecting to server...", m.spinner.View())}
		} else {
			secondsLeft := int(time.Until(m.nextRetryTime).Seconds())
			if secondsLeft < 0 {
				secondsLeft = 0
			}
			lines = []string{fmt.Sprintf("%s Connecting to server in %d seconds...", m.spinner.View(), secondsLeft)}
		}
		return m.centerContent(lines)
	}

	bannerColors := []string{"#2d6e2d", "#256025", "#1d531d", "#154515", "#0d380d", "#071d07"}
	bannerText := []string{
		"███╗   ███╗██╗████████╗██████╗  █████╗ ",
		"████╗ ████║██║╚══██╔══╝██╔══██╗██╔══██╗",
		"██╔████╔██║██║   ██║   ██████╔╝███████║",
		"██║╚██╔╝██║██║   ██║   ██╔══██╗██╔══██║",
		"██║ ╚═╝ ██║██║   ██║   ██║  ██║██║  ██║",
		"╚═╝     ╚═╝╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝",
	}

	var lines []string

	// Banner: horizontally centered, with top padding
	for i := 0; i < bannerTopPad; i++ {
		lines = append(lines, "")
	}
	for i, text := range bannerText {
		leftPad := (m.width - runewidth.StringWidth(text)) / 2
		if leftPad < 0 {
			leftPad = 0
		}
		styled := lipgloss.NewStyle().Foreground(lipgloss.Color(bannerColors[i])).Render(text)
		lines = append(lines, strings.Repeat(" ", leftPad)+styled)
	}
	lines = append(lines, "")

	// Content: left-padded
	var contentLines []string
	if m.reposListReady {
		contentLines = append(contentLines, strings.Split(m.reposList.View(), "\n")...)
	}
	if m.sessionsListReady {
		contentLines = append(contentLines, strings.Split(m.sessionsList.View(), "\n")...)
	}
	if m.focusedPane == paneSessions {
		contentLines = append(contentLines, "Tab: switch pane | ↵: attach | q: quit")
	} else {
		contentLines = append(contentLines, "Tab: switch pane | q: quit")
	}
	for _, line := range contentLines {
		lines = append(lines, "  "+line)
	}

	return strings.Join(lines, "\n")
}

func (m DashboardModel) getWorktreesForRepo(repoID string) []*proto.Worktree {
	var result []*proto.Worktree
	for _, wt := range m.worktrees {
		if wt.RepoId == repoID {
			result = append(result, wt)
		}
	}
	return result
}

func (m DashboardModel) centerContent(lines []string) string {
	contentHeight := len(lines)

	var centered []string
	for _, line := range lines {
		lineWidth := runewidth.StringWidth(line)
		leftPadding := (m.width - lineWidth) / 2
		if leftPadding < 0 {
			leftPadding = 0
		}
		centered = append(centered, strings.Repeat(" ", leftPadding)+line)
	}

	topPadding := (m.height - contentHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	result := strings.Repeat("\n", topPadding) + strings.Join(centered, "\n")
	return result
}

func RunDashboard() error {
	p := tea.NewProgram(NewDashboard(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
