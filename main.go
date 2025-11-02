package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/shirou/gopsutil/cpu" // 用于获取 CPU 详细信息
	"github.com/shirou/gopsutil/host" // 用于获取系统启动时间
	"github.com/shirou/gopsutil/mem" // 用于获取内存信息
	"github.com/shirou/gopsutil/disk" // 用于获取磁盘信息
)

// 定义一个结构体来保存和更新UI元素
type SystemInfoUI struct {
	hostnameLabel  *widget.Label
	osLabel        *widget.Label
	archLabel      *widget.Label
	cpuNameLabel   *widget.Label
	cpuCoresLabel  *widget.Label
	memTotalLabel  *widget.Label
	memUsedLabel   *widget.Label
	memFreeLabel   *widget.Label
	diskTotalLabel *widget.Label
	diskUsedLabel  *widget.Label
	diskFreeLabel  *widget.Label
	userLabel      *widget.Label
	uptimeLabel    *widget.Label
	timeLabel      *widget.Label
}

func main() {
	a := app.New()
	w := a.NewWindow("系统信息查看器")
	w.Resize(fyne.NewSize(500, 400)) // 设置窗口初始大小

	ui := &SystemInfoUI{
		hostnameLabel:  widget.NewLabel(""),
		osLabel:        widget.NewLabel(""),
		archLabel:      widget.NewLabel(""),
		cpuNameLabel:   widget.NewLabel(""),
		cpuCoresLabel:  widget.NewLabel(""),
		memTotalLabel:  widget.NewLabel(""),
		memUsedLabel:   widget.NewLabel(""),
		memFreeLabel:   widget.NewLabel(""),
		diskTotalLabel: widget.NewLabel(""),
		diskUsedLabel:  widget.NewLabel(""),
		diskFreeLabel:  widget.NewLabel(""),
		userLabel:      widget.NewLabel(""),
		uptimeLabel:    widget.NewLabel(""),
		timeLabel:      widget.NewLabel(""),
	}

	// 使用 Form 布局来美化信息的展示
	form := &widget.Form{
		Items: []*widget.FormItem{
			widget.NewFormItem("电脑名称", ui.hostnameLabel),
			widget.NewFormItem("操作系统", ui.osLabel),
			widget.NewFormItem("系统架构", ui.archLabel),
			widget.NewFormItem("CPU", ui.cpuNameLabel),
			widget.NewFormItem("CPU 核心数", ui.cpuCoresLabel),
			widget.NewFormItem("内存总量", ui.memTotalLabel),
			widget.NewFormItem("已用内存", ui.memUsedLabel),
			widget.NewFormItem("剩余内存", ui.memFreeLabel),
			widget.NewFormItem("磁盘总量", ui.diskTotalLabel),
			widget.NewFormItem("已用磁盘", ui.diskUsedLabel),
			widget.NewFormItem("剩余磁盘", ui.diskFreeLabel),
			widget.NewFormItem("当前用户", ui.userLabel),
			widget.NewFormItem("系统运行时间", ui.uptimeLabel),
			widget.NewFormItem("当前时间", ui.timeLabel),
		},
	}

	// 将 Form 放到一个可滚动的容器中，以防信息过多超出窗口
	scrollContainer := container.NewVScroll(form)
	
	// 更新信息的函数
	updateInfo := func() {
		updateSystemInfo(ui)
	}

	// 首次更新信息
	updateInfo()

	// 定时更新信息 (每5秒)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			updateInfo()
		}
	}()

	w.SetContent(container.New(layout.NewMaxLayout(), scrollContainer))
	w.ShowAndRun()
}

// updateSystemInfo 收集并更新UI上的所有系统信息
func updateSystemInfo(ui *SystemInfoUI) {
	// 主机名
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "获取失败"
	}
	ui.hostnameLabel.SetText(hostname)

	// OS 和 架构
	ui.osLabel.SetText(runtime.GOOS)
	ui.archLabel.SetText(runtime.GOARCH)

	// CPU 信息
	cpuInfos, err := cpu.Info()
	if err == nil && len(cpuInfos) > 0 {
		ui.cpuNameLabel.SetText(cpuInfos[0].ModelName)
		ui.cpuCoresLabel.SetText(fmt.Sprintf("%d (逻辑核: %d)", cpuInfos[0].Cores, runtime.NumCPU()))
	} else {
		ui.cpuNameLabel.SetText("获取失败")
		ui.cpuCoresLabel.SetText(fmt.Sprintf("%d", runtime.NumCPU()))
	}
	
	// 内存信息
	vMem, err := mem.VirtualMemory()
	if err == nil {
		ui.memTotalLabel.SetText(fmt.Sprintf("%.2f GB", byteToGB(vMem.Total)))
		ui.memUsedLabel.SetText(fmt.Sprintf("%.2f GB (%.2f%%)", byteToGB(vMem.Used), vMem.UsedPercent))
		ui.memFreeLabel.SetText(fmt.Sprintf("%.2f GB", byteToGB(vMem.Free)))
	} else {
		ui.memTotalLabel.SetText("获取失败")
		ui.memUsedLabel.SetText("获取失败")
		ui.memFreeLabel.SetText("获取失败")
	}

	// 磁盘信息 (只取根分区或第一个分区)
	partitions, err := disk.Partitions(false) // false表示不包含CD-ROM等
	if err == nil && len(partitions) > 0 {
		usage, err := disk.Usage(partitions[0].Mountpoint)
		if err == nil {
			ui.diskTotalLabel.SetText(fmt.Sprintf("%.2f GB", byteToGB(usage.Total)))
			ui.diskUsedLabel.SetText(fmt.Sprintf("%.2f GB (%.2f%%)", byteToGB(usage.Used), usage.UsedPercent))
			ui.diskFreeLabel.SetText(fmt.Sprintf("%.2f GB", byteToGB(usage.Free)))
		} else {
			ui.diskTotalLabel.SetText("获取失败")
			ui.diskUsedLabel.SetText("获取失败")
			ui.diskFreeLabel.SetText("获取失败")
		}
	} else {
		ui.diskTotalLabel.SetText("获取失败")
		ui.diskUsedLabel.SetText("获取失败")
		ui.diskFreeLabel.SetText("获取失败")
	}

	// 当前用户 (Go标准库os/user可以获取，但可能在交叉编译时依赖CGO，
	// 为了简化这里暂时用os.Getenv获取，实际应用推荐gopsutil/process)
	currentUser := os.Getenv("USERNAME") // Windows
	if currentUser == "" {
		currentUser = os.Getenv("USER") // Linux/macOS
	}
	if currentUser == "" {
		currentUser = "未知"
	}
	ui.userLabel.SetText(currentUser)

	// 系统运行时间 (Uptime)
	uptime, err := host.Uptime()
	if err == nil {
		ui.uptimeLabel.SetText(formatDuration(time.Duration(uptime) * time.Second))
	} else {
		ui.uptimeLabel.SetText("获取失败")
	}

	// 当前时间
	ui.timeLabel.SetText(time.Now().Format("2006-01-02 15:04:05"))
	
	// 强制UI刷新 (重要，确保定时更新生效)
	fyne.CurrentApp().Driver().CallOnMainThread(func() {
		ui.hostnameLabel.Refresh()
		ui.osLabel.Refresh()
		ui.archLabel.Refresh()
		ui.cpuNameLabel.Refresh()
		ui.cpuCoresLabel.Refresh()
		ui.memTotalLabel.Refresh()
		ui.memUsedLabel.Refresh()
		ui.memFreeLabel.Refresh()
		ui.diskTotalLabel.Refresh()
		ui.diskUsedLabel.Refresh()
		ui.diskFreeLabel.Refresh()
		ui.userLabel.Refresh()
		ui.uptimeLabel.Refresh()
		ui.timeLabel.Refresh()
	})
}

// 辅助函数：字节转换为 GB
func byteToGB(b uint64) float64 {
	return float64(b) / (1024 * 1024 * 1024)
}

// 辅助函数：格式化时间段
func formatDuration(d time.Duration) string {
    days := int(d.Hours() / 24)
    hours := int(d.Hours()) % 24
    minutes := int(d.Minutes()) % 60
    seconds := int(d.Seconds()) % 60
    return fmt.Sprintf("%d天 %02d小时 %02d分 %02d秒", days, hours, minutes, seconds)
}