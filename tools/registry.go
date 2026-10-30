package tools

// GetAllTools 返回所有可用工具，供 Agent 使用。
func GetAllTools() map[string]*ToolDef {
	all := []*ToolDef{
		NewGetCurrentTime(),
		NewRecordEvent(),
		NewDaysUntil(),
		NewDeleteEvent(), 
		NewUpdateEvent(),
	}

	m := make(map[string]*ToolDef, len(all))
	for _, t := range all {
		m[t.Name] = t
	}
	return m
}