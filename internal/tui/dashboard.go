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
	"mitra/internal/util"
)

type dataLoadedMsg struct {
	repos     []*proto.Repo
	worktrees []*proto.Worktree
	err       error
}

type tickMsg time.Time

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
		if wt.IsMain {
			branches = append(branches, fmt.Sprintf("★ %s", wt.Branch))
		} else {
			branches = append(branches, wt.Branch)
		}
	}

	if len(branches) > 3 {
		return fmt.Sprintf("%d worktrees: %s, ...", len(branches), strings.Join(branches[:3], ", "))
	}
	return fmt.Sprintf("%d worktrees: %s", len(branches), strings.Join(branches, ", "))
}

func (i repoItem) FilterValue() string {
	return i.Title()
}

type DashboardModel struct {
	width          int
	height         int
	quitting       bool
	repos          []*proto.Repo
	worktrees      []*proto.Worktree
	reposList      list.Model
	reposListReady bool
	err            error
	spinner        spinner.Model
	connecting     bool
	nextRetryTime  time.Time
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
		if m.reposListReady {
			m.reposList.SetSize(msg.Width-4, msg.Height-9)
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		default:
			if m.reposListReady && len(m.repos) > 0 {
				var cmd tea.Cmd
				m.reposList, cmd = m.reposList.Update(msg)
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

		items := make([]list.Item, len(m.repos))
		for i, repo := range m.repos {
			items[i] = repoItem{
				repo:      repo,
				worktrees: m.getWorktreesForRepo(repo.Id),
			}
		}

		m.reposList = list.New(items, list.NewDefaultDelegate(), m.width-4, m.height-9)
		m.reposList.Title = "Repositories"
		m.reposList.SetShowStatusBar(false)
		m.reposList.SetFilteringEnabled(true)
		m.reposListReady = true
		return m, nil
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

	var lines []string
	lines = append(lines, "███╗   ███╗██╗████████╗██████╗  █████╗ ")
	lines = append(lines, "████╗ ████║██║╚══██╔══╝██╔══██╗██╔══██╗")
	lines = append(lines, "██╔████╔██║██║   ██║   ██████╔╝███████║")
	lines = append(lines, "██║╚██╔╝██║██║   ██║   ██╔══██╗██╔══██║")
	lines = append(lines, "██║ ╚═╝ ██║██║   ██║   ██║  ██║██║  ██║")
	lines = append(lines, "╚═╝     ╚═╝╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝")
	lines = append(lines, "")

	if len(m.repos) == 0 {
		lines = append(lines, "No repositories found")
		lines = append(lines, "")
		lines = append(lines, "Add a repo with: mitra repo add <url>")
		lines = append(lines, "")
		lines = append(lines, "q quit")
	} else {
		listView := m.reposList.View()
		lines = append(lines, strings.Split(listView, "\n")...)
	}

	paddedLines := make([]string, len(lines))
	for i, line := range lines {
		paddedLines[i] = "  " + line
	}

	return strings.Join(paddedLines, "\n")
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
