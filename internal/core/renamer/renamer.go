package renamer

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// MagicVariables 预定义的魔法正则变量
// 注意：Go 的 regexp 不支持 Lookaround (断言)，需使用 \b 或捕获组
var MagicVariables = map[string]string{
	"{YEAR}":    `\b(?:18|19|20)\d{2}\b`,
	"{DATE}":    `\b(?:18|19|20)?\d{2}[\.\-/年]\d{1,2}[\.\-/月]\d{1,2}\b`,
	"{CHINESE}": `[\u4e00-\u9fa5]{2,}`,
	"{EXT}":     `\.(\w+)$`, // 使用捕获组提取后缀
}

// RenameOptions 重命名选项
type RenameOptions struct {
	TaskName    string
	Pattern     string // 用户定义的原始正则匹配式
	Replacement string // 用户定义的替换模板 (含变量或 Go template)
	FileName    string // 原始文件名
}

// Processor 重命名处理器
type Processor struct{}

func NewProcessor() *Processor {
	return &Processor{}
}

// Process 执行重命名逻辑
func (p *Processor) Process(opts RenameOptions) (string, error) {
	if opts.Replacement == "" {
		return opts.FileName, nil
	}

	result := opts.Replacement

	// 1. 替换基础变量 {TASKNAME}
	result = strings.ReplaceAll(result, "{TASKNAME}", opts.TaskName)

	// 2. 尝试从原文件名中通过正则提取魔法变量的值并替换到 result 中
	for varName, regPattern := range MagicVariables {
		if strings.Contains(result, varName) {
			re := regexp.MustCompile(regPattern)
			matches := re.FindStringSubmatch(opts.FileName)
			if len(matches) > 0 {
				// 如果正则中有捕获组（如 {EXT}），则取第一个捕获组的内容
				// 否则取整个匹配到的字符串内容
				match := matches[0]
				if len(matches) > 1 {
					match = matches[1]
				}

				// 特殊处理日期格式
				if varName == "{DATE}" {
					match = p.cleanDate(match)
				}
				result = strings.ReplaceAll(result, varName, match)
			} else {
				// 未匹配到则置空
				result = strings.ReplaceAll(result, varName, "")
			}
		}
	}

	// 3. 执行正则子组替换 (如果 Pattern 和 Replacement 同时存在)
	if opts.Pattern != "" && strings.Contains(result, "$") {
		re, err := regexp.Compile(opts.Pattern)
		if err == nil {
			result = re.ReplaceAllString(opts.FileName, result)
		}
	}

	// 4. 执行 Go Template 动态渲染 (高级模式)
	if strings.Contains(result, "{{") {
		tmpl, err := template.New("rename").Parse(result)
		if err == nil {
			var buf bytes.Buffer
			data := map[string]interface{}{
				"TaskName": opts.TaskName,
				"OldName":  opts.FileName,
				"Now":      time.Now(),
			}
			if err := tmpl.Execute(&buf, data); err == nil {
				result = buf.String()
			}
		}
	}

	return strings.TrimSpace(result), nil
}

func (p *Processor) cleanDate(input string) string {
	// 移除非数字字符，统一为 YYYYMMDD 或 YYMMDD
	re := regexp.MustCompile(`\D`)
	return re.ReplaceAllString(input, "")
}
