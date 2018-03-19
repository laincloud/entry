package models

import (
	"bytes"
	"regexp"
	"text/template"
	"time"

	swaggermodels "github.com/laincloud/entry/server/gen/models"
	"github.com/laincloud/entry/server/global"
	"github.com/laincloud/entry/server/util"
)

const (
	mailTemplate = `Subject: [Entry@{{.LAINDomain}}][{{.Command.Session.AppName}}] - Dangerous Command
MIME-version: 1.0;
Content-Type: text/html; charset="UTF-8";

<html lang="en">

<head>
    <style type="text/css">
        table {
            border-collapse: collapse;
        }

        caption {
            margin: 1em;
        }

        td,
        th {
            border: 1px solid #cccccc;
            padding: 0.6em;
        }

        tr:nth-child(even) {
            background-color: #dddddd;
        }
    </style>
</head>

<body>
    <div style="margin-top: 2em; margin-left: 1em">
        <table>
            <caption>Command</caption>
            <tr>
                <td>Command</td>
                <td>{{.Command.Content}}</td>
            </tr>

            <tr>
                <td>Created At</td>
                <td>{{.Command.CreatedAt}}</td>
            </tr>

            <tr>
                <td>Command ID</td>
                <td>{{.Command.CommandID}}</td>
            </tr>
        </table>

        <table style="margin-top: 2em">
            <caption>Additional Infomation</caption>
            <tr>
                <td>App Name</td>
                <td>{{.Command.Session.AppName}}</td>
            </tr>

            <tr>
                <td>User</td>
                <td>{{.Command.Session.User}}</td>
            </tr>

            <tr>
                <td>Source IP</td>
                <td>{{.Command.Session.SourceIP}}</td>
            </tr>

            <tr>
                <td>Proc Name</td>
                <td>{{.Command.Session.ProcName}}</td>
            </tr>

            <tr>
                <td>Instance No</td>
                <td>{{.Command.Session.InstanceNo}}</td>
            </tr>

            <tr>
                <td>Node IP</td>
                <td>{{.Command.Session.NodeIP}}</td>
            </tr>

            <tr>
                <td>Session ID</td>
                <td>{{.Command.SessionID}}</td>
            </tr>
        </table>
    </div>
</body>

</html>
`
)

var (
	riskyCommandPatterns = []string{
		"vipw",                      // 编辑用户密码文件
		"ettercap",                  // 嗅探
		"chmod\\s+777",              // 修改权限
		"useradd",                   // 添加用户
		"(vim|vi)\\s+mysql_history", // 修改mysql日志
		"cat\\s+/etc/passwd",        // 查看系统用户
		"nmap",                      // nmap扫描
		"arpspoof",                  // arp欺骗
		"lcx",                       // 使用代理软件
		"rcsocks",                   // socks反弹代理
		"bash\\s+-i",                // 反弹shell
		"history\\s+-c",             // 清除日志记录
		"exec",                      // 反弹
		"unset\\s+HISTORY",          // 不记录历史命令
		"portmap",                   // 端口转发
		"export\\s+HISTSIZE=0",      // 设置操作命令不记录进日志
		"rm\\s+-rf\\s+/",            // 强制删除根目录
		"halt",                      // 关机
		"poweroff",                  // 关机
		"shutdown",                  //关机
	}
)

// Command denotes the command typed by user
type Command struct {
	CommandID int64   `gorm:"primary_key"`
	Session   Session `gorm:"foreignkey:SessionID;association_foreignkey:SessionID"`
	SessionID int64
	User      string `gorm:"index"`
	Content   string
	CreatedAt time.Time `sql:"not null;DEFAULT:current_timestamp"`
}

// SwaggerModel return the swagger version
func (c Command) SwaggerModel() swaggermodels.Command {
	return swaggermodels.Command{
		CommandID:  c.CommandID,
		User:       c.User,
		AppName:    c.Session.AppName,
		ProcName:   c.Session.ProcName,
		InstanceNo: c.Session.InstanceNo,
		Content:    c.Content,
		SessionID:  c.SessionID,
		CreatedAt:  c.CreatedAt.Unix(),
	}
}

// IsRisky judge whether this command is risky
func (c Command) IsRisky() bool {
	return isRisky(c.Content)
}

func isRisky(commandContent string) bool {
	for _, p := range riskyCommandPatterns {
		if matched, _ := regexp.MatchString(p, commandContent); matched {
			return true
		}
	}

	return false
}

// MailData will be inserted into mailTemplate
type MailData struct {
	Command    Command
	LAINDomain string
}

// Alert alert dangerous command
func (c Command) Alert(g *global.Global) error {
	msg, err := c.newMailMessage(g.LAINDomain)
	if err != nil {
		return err
	}

	return util.SendMail(msg, g)
}

func (c Command) newMailMessage(lainDomain string) ([]byte, error) {
	t, err := template.New("mail").Parse(mailTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	data := MailData{
		Command:    c,
		LAINDomain: lainDomain,
	}
	if err = t.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
