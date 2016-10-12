package client

import (
	"time"

	"github.com/SDHM/sqlregret/binlogevent"
	"github.com/SDHM/sqlregret/config"
	"github.com/SDHM/sqlregret/protocol"
)

// 返回为false 表示不过滤 返回为true 表示过滤
func FilterSkipSQL(eventType int) bool {

	// 如果eventType不是这六个之间的一个, 则不过滤
	if eventType != binlogevent.UPDATE_ROWS_EVENT_V1 &&
		eventType != binlogevent.UPDATE_ROWS_EVENT &&
		eventType != binlogevent.DELETE_ROWS_EVENT_V1 &&
		eventType != binlogevent.DELETE_ROWS_EVENT &&
		eventType != binlogevent.WRITE_ROWS_EVENT_V1 &&
		eventType != binlogevent.WRITE_ROWS_EVENT {
		return false
	}

	switch config.G_filterConfig.FilterSQL {
	case "":
		{
			return false
		}
	case "insert":
		{
			if eventType == binlogevent.UPDATE_ROWS_EVENT_V1 ||
				eventType == binlogevent.UPDATE_ROWS_EVENT ||
				eventType == binlogevent.DELETE_ROWS_EVENT_V1 ||
				eventType == binlogevent.DELETE_ROWS_EVENT {
				g_transaction.SkipSomeThing()
				return true
			} else if eventType == binlogevent.WRITE_ROWS_EVENT_V1 ||
				eventType == binlogevent.WRITE_ROWS_EVENT {
				return false
			}
		}
	case "delete":
		{
			if eventType == binlogevent.UPDATE_ROWS_EVENT_V1 ||
				eventType == binlogevent.UPDATE_ROWS_EVENT ||
				eventType == binlogevent.WRITE_ROWS_EVENT_V1 ||
				eventType == binlogevent.WRITE_ROWS_EVENT {
				g_transaction.SkipSomeThing()
				return true
			} else if eventType == binlogevent.DELETE_ROWS_EVENT_V1 ||
				eventType == binlogevent.DELETE_ROWS_EVENT {
				return false
			}
		}
	case "update":
		{
			if eventType == binlogevent.DELETE_ROWS_EVENT_V1 ||
				eventType == binlogevent.DELETE_ROWS_EVENT ||
				eventType == binlogevent.WRITE_ROWS_EVENT_V1 ||
				eventType == binlogevent.WRITE_ROWS_EVENT {
				g_transaction.SkipSomeThing()
				return true
			} else if eventType == binlogevent.UPDATE_ROWS_EVENT_V1 ||
				eventType == binlogevent.UPDATE_ROWS_EVENT {
				return false
			}
		}
	default:
		{
			return true
		}
	}

	return false
}

func FilterMode(eventType int) bool {
	if config.G_filterConfig.Mode == "mark" {
		if eventType == binlogevent.FORMAT_DESCRIPTION_EVENT ||
			eventType == binlogevent.TABLE_MAP_EVENT ||
			eventType == binlogevent.ROTATE_EVENT {
			return false
		} else {
			return true
		}
	} else if config.G_filterConfig.Mode == "parse" {
		return false
	} else {
		return false
	}
}

func FilterTime(timeSnap time.Time, eventType int) bool {
	if config.G_filterConfig.StartTimeEnable() && config.G_filterConfig.EndTimeEnable() {
		//开始时间和结束时间都设置了
		if timeSnap.After(config.G_filterConfig.StartTime) && timeSnap.Before(config.G_filterConfig.EndTime) {

			return false
		} else {
			//时间在两者之外，并且不是修改操作的直接跳过
			if eventType == binlogevent.WRITE_ROWS_EVENT_V1 || eventType == binlogevent.WRITE_ROWS_EVENT ||
				eventType == binlogevent.UPDATE_ROWS_EVENT_V1 || eventType == binlogevent.UPDATE_ROWS_EVENT ||
				eventType == binlogevent.DELETE_ROWS_EVENT_V1 || eventType == binlogevent.DELETE_ROWS_EVENT {
				return true
			} else {
				return false
			}
		}
	} else if config.G_filterConfig.StartTimeEnable() && !config.G_filterConfig.EndTimeEnable() {
		if timeSnap.After(config.G_filterConfig.StartTime) {
			return false
		} else {
			return true
		}
	}

	return false
}

func FilterPos(eventType int, fileIndex int, pos int64) bool {

	if config.G_filterConfig.StartPosEnable() && config.G_filterConfig.EndPosEnable() {
		if (fileIndex >= config.G_filterConfig.StartFileIndex && int(pos) >= config.G_filterConfig.StartPos) &&
			(fileIndex <= config.G_filterConfig.EndFileIndex && int(pos) <= config.G_filterConfig.EndPos) {
			return false
		} else {
			if eventType == binlogevent.WRITE_ROWS_EVENT_V1 || eventType == binlogevent.WRITE_ROWS_EVENT ||
				eventType == binlogevent.UPDATE_ROWS_EVENT_V1 || eventType == binlogevent.UPDATE_ROWS_EVENT ||
				eventType == binlogevent.DELETE_ROWS_EVENT_V1 || eventType == binlogevent.DELETE_ROWS_EVENT {
				return true
			} else {
				return false
			}
		}
	} else if config.G_filterConfig.StartPosEnable() && !config.G_filterConfig.EndTimeEnable() {
		if fileIndex >= config.G_filterConfig.StartFileIndex && int(pos) >= config.G_filterConfig.StartPos {
			return false
		} else {
			if eventType == binlogevent.WRITE_ROWS_EVENT_V1 || eventType == binlogevent.WRITE_ROWS_EVENT ||
				eventType == binlogevent.UPDATE_ROWS_EVENT_V1 || eventType == binlogevent.UPDATE_ROWS_EVENT ||
				eventType == binlogevent.DELETE_ROWS_EVENT_V1 || eventType == binlogevent.DELETE_ROWS_EVENT {
				return true
			} else {
				return false
			}
		}
	} else if !config.G_filterConfig.StartPosEnable() && !config.G_filterConfig.EndTimeEnable() {
		return false
	}

	return true
}

func FilterColumns(eventType protocol.EventType, tableMeta *TableMeta, columns []*Column) bool {
	if (eventType == protocol.EventType_INSERT && config.G_filterConfig.WithInsertFilterColumn()) ||
		(eventType == protocol.EventType_UPDATE && config.G_filterConfig.WithUpdateFilterColumn()) {

		findColumn := false
		var filterColumns []*config.ColumnFilter
		if eventType == protocol.EventType_INSERT {
			filterColumns = config.G_filterConfig.InsertFilterColumn
		}

		if eventType == protocol.EventType_UPDATE {
			filterColumns = config.G_filterConfig.UpdateFilterColumn
		}

		for i, _ := range columns {
			if nil != tableMeta {
				columnName := tableMeta.Fileds[i].ColumnName

				for _, c := range filterColumns {
					if c.GetName() == columnName {
						findColumn = true
						goto Out1
					}
				}
			}
		}

	Out1:
		return !findColumn
	}

	return false
}