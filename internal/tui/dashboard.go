package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"mitra/internal/config"
	"mitra/internal/proto"
	"mitra/internal/util"
)

type viewState int

const (
	stateWelcome viewState = iota
	stateDashboard
)

type tabView int

const (
	tabOverview tabView = iota
	tabRepos
)

type dataLoadedMsg struct {
	repos     []*proto.Repo
	worktrees []*proto.Worktree
	err       error
}

type DashboardModel struct {
	width     int
	height    int
	quitting  bool
	state     viewState
	activeTab tabView
	repos     []*proto.Repo
	worktrees []*proto.Worktree
	loading   bool
	err       error
}

func NewDashboard() DashboardModel {
	return DashboardModel{
		state: stateWelcome,
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return nil
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

	return dataLoadedMsg{
		repos:     repoResp.Repos,
		worktrees: worktreeResp.Worktrees,
	}
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "tab", "right":
			if m.state == stateDashboard {
				m.activeTab = (m.activeTab + 1) % 2
			}
			return m, nil
		case "shift+tab", "left":
			if m.state == stateDashboard {
				m.activeTab = (m.activeTab - 1 + 2) % 2
			}
			return m, nil
		default:
			if m.state == stateWelcome && !m.loading {
				m.state = stateDashboard
				m.loading = true
				return m, loadData
			}
		}
	case dataLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.repos = msg.repos
			m.worktrees = msg.worktrees
		}
		return m, nil
	}

	return m, nil
}

func (m DashboardModel) View() string {
	if m.quitting {
		return ""
	}

	switch m.state {
	case stateWelcome:
		return m.renderWelcome()
	case stateDashboard:
		return m.renderDashboard()
	}

	return ""
}

func (m DashboardModel) renderTabs() string {
	tabs := []string{"Overview", "Repos"}
	var tabStrs []string

	for i, tab := range tabs {
		if tabView(i) == m.activeTab {
			tabStrs = append(tabStrs, fmt.Sprintf("[ %s ]", tab))
		} else {
			tabStrs = append(tabStrs, fmt.Sprintf("  %s  ", tab))
		}
	}

	return strings.Join(tabStrs, " ")
}

func (m DashboardModel) renderWelcome() string {
	lines := []string{
		"███╗   ███╗██╗████████╗██████╗  █████╗ ",
		"████╗ ████║██║╚══██╔══╝██╔══██╗██╔══██╗",
		"██╔████╔██║██║   ██║   ██████╔╝███████║",
		"██║╚██╔╝██║██║   ██║   ██╔══██╗██╔══██║",
		"██║ ╚═╝ ██║██║   ██║   ██║  ██║██║  ██║",
		"╚═╝     ╚═╝╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝",
		"",
		"Repos, Worktrees, Branches and Agents",
		"",
		"",
		"Press any key to continue or q to quit",
	}

	return m.centerContent(lines)
}

func (m DashboardModel) renderDashboard() string {
	if m.loading {
		lines := []string{"Loading..."}
		return m.centerContent(lines)
	}

	if m.err != nil {
		lines := []string{
			"Error loading data:",
			m.err.Error(),
			"",
			"Press q to quit",
		}
		return m.centerContent(lines)
	}

	var lines []string
	lines = append(lines, "")
	lines = append(lines, "MITRA DASHBOARD")
	lines = append(lines, "")
	lines = append(lines, m.renderTabs())
	lines = append(lines, "")

	switch m.activeTab {
	case tabOverview:
		lines = append(lines, m.renderOverviewTab()...)
	case tabRepos:
		lines = append(lines, m.renderReposTab()...)
	}

	lines = append(lines, "")
	lines = append(lines, "Tab/→ next | Shift+Tab/← prev | q quit")

	// Add left padding
	paddedLines := make([]string, len(lines))
	for i, line := range lines {
		paddedLines[i] = "  " + line
	}

	return strings.Join(paddedLines, "\n")
}

func (m DashboardModel) renderOverviewTab() []string {
	var lines []string

	repoCount := len(m.repos)
	worktreeCount := len(m.worktrees)

	lines = append(lines, "═══════════════════════════════════════")
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Repositories: %d", repoCount))
	lines = append(lines, fmt.Sprintf("Worktrees:    %d", worktreeCount))
	lines = append(lines, "")
	lines = append(lines, "═══════════════════════════════════════")

	return lines
}

func (m DashboardModel) renderReposTab() []string {
	var lines []string

	if len(m.repos) == 0 {
		lines = append(lines, "No repositories found")
		lines = append(lines, "")
		lines = append(lines, "Add a repo with: mitra repo add <url>")
		return lines
	}

	for i, repo := range m.repos {
		lines = append(lines, fmt.Sprintf("%d. %s/%s", i+1, repo.Owner, repo.Repo))

		repoWorktrees := m.getWorktreesForRepo(repo.Id)
		if len(repoWorktrees) == 0 {
			lines = append(lines, "   └─ No worktrees")
		} else {
			for j, wt := range repoWorktrees {
				isLast := j == len(repoWorktrees)-1
				prefix := "├─"
				if isLast {
					prefix = "└─"
				}

				indicator := ""
				if wt.IsMain {
					indicator = " ★"
				}

				lines = append(lines, fmt.Sprintf("   %s %s%s", prefix, wt.Branch, indicator))
			}
		}
		lines = append(lines, "")
	}

	return lines
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
