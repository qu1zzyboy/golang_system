package bootx

import (
	"context"
	"fmt"

	"github.com/hhh500/quantGoInfra/infra/errorx"
)

type BootManager struct {
	components map[string]Bootable // 注册的所有组件
	started    []Bootable          // 启动成功的组件(用于逆序停止)
}

func (bm *BootManager) Register(comp Bootable) {
	bm.components[comp.ModuleId()] = comp
}

func (bm *BootManager) StartAll(ctx context.Context) {
	visited := map[string]bool{}  // 标记已经成功启动的组件
	visiting := map[string]bool{} // 标记正在启动的组件(用于检测循环依赖)

	// 深度优先搜索启动组件
	var dfs func(string) error
	dfs = func(moduleId string) error {
		if visited[moduleId] {
			return nil
		}
		if visiting[moduleId] {
			return fmt.Errorf("存在循环依赖: %s", moduleId)
		}
		comp, exists := bm.components[moduleId]
		if !exists {
			errorx.PanicWithCaller(fmt.Sprintf("模块 %s 未注册", moduleId))
		}
		visiting[moduleId] = true

		// 先递归启动依赖的组件
		for _, dep := range comp.DependsOn() {
			if err := dfs(dep); err != nil {
				return err
			}
		}
		visiting[moduleId] = false

		// 启动当前组件
		if err := comp.Start(ctx); err != nil {
			errorx.PanicWithCaller(fmt.Sprintf("模块 %s 启动失败: %s", moduleId, err))
		}
		visited[moduleId] = true
		bm.started = append(bm.started, comp)
		return nil
	}

	// 遍历所有组件,启动它们
	for name := range bm.components {
		dfs(name)
	}
}

func (bm *BootManager) StopAll(ctx context.Context) error {
	for i := len(bm.started) - 1; i >= 0; i-- {
		comp := bm.started[i]
		fmt.Printf("<< Stopping component: %s\n", comp.ModuleId())
		if err := comp.Stop(ctx); err != nil {
			return fmt.Errorf("stop failed: %s: %w", comp.ModuleId(), err)
		}
	}
	return nil
}
