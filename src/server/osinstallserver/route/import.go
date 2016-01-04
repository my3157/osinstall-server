package route

import (
	"fmt"
	"github.com/AlexanderChen1989/go-json-rest/rest"
	"golang.org/x/net/context"
	"middleware"
	//"server/osinstallserver/util"
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"github.com/qiniu/iconv"
	"io"
	"os"
	"regexp"
	"server/osinstallserver/util"
	"strings"
	"time"
)

func UploadDevice(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	r.ParseForm()
	file, handle, err := r.FormFile("file")
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	cd, err := iconv.Open("UTF-8", "GBK")
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}
	defer cd.Close()

	dir := "./upload/"
	if !util.FileExist(dir) {
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}

	list := strings.Split(handle.Filename, ".")
	fix := list[len(list)-1]

	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%s", time.Now().UnixNano()) + handle.Filename))
	cipherStr := h.Sum(nil)
	md5 := fmt.Sprintf("%s", hex.EncodeToString(cipherStr))
	filename := md5 + "." + fix

	result := make(map[string]interface{})
	result["result"] = filename

	if util.FileExist(dir + filename) {
		os.Remove(dir + filename)
	}

	f, err := os.OpenFile(dir+filename, os.O_WRONLY|os.O_CREATE, 0666)
	io.Copy(f, file)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	defer f.Close()
	defer file.Close()
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
	return
}

func ImportPriview(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		Filename string
		Limit    uint
		Offset   uint
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	file := "./upload/" + info.Filename

	cd, err := iconv.Open("utf-8", "gbk") // convert gbk to utf8
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}
	defer cd.Close()

	input, err := os.Open(file)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}
	bufSize := 1024 * 1024
	read := iconv.NewReader(cd, input, bufSize)

	r2 := csv.NewReader(read)
	ra, err := r2.ReadAll()
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	length := len(ra)

	type Device struct {
		ID              uint
		BatchNumber     string
		Sn              string
		Hostname        string
		Ip              string
		NetworkID       uint
		OsID            uint
		HardwareID      uint
		SystemID        uint
		Location        string
		LocationID      uint
		AssetNumber     string
		Status          string
		InstallProgress float64
		InstallLog      string
		NetworkName     string
		OsName          string
		HardwareName    string
		SystemName      string
		Content         string
	}
	var success []Device
	var failure []Device
	//var result []string
	for i := 1; i < length; i++ {
		//result = append(result, ra[i][0])
		var row Device
		if len(ra[i]) != 8 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "导入文件格式错误!"
			failure = append(failure, row)
			continue
		}

		row.Sn = strings.TrimSpace(ra[i][0])
		row.Hostname = strings.TrimSpace(ra[i][1])
		row.Ip = strings.TrimSpace(ra[i][2])
		row.OsName = strings.TrimSpace(ra[i][3])
		row.HardwareName = strings.TrimSpace(ra[i][4])
		row.SystemName = strings.TrimSpace(ra[i][5])
		row.Location = strings.TrimSpace(ra[i][6])
		row.AssetNumber = strings.TrimSpace(ra[i][7])

		if len(row.Sn) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "SN长度超过255限制!"
		}

		if len(row.Hostname) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "主机名长度超过255限制!"
		}

		if len(row.Location) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "位置长度超过255限制!"
		}

		if len(row.AssetNumber) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "财编长度超过255限制!"
		}

		if row.Sn == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "SN不能为空!"
		}

		if row.Hostname == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "主机名不能为空!"
		}

		if row.Ip == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "IP不能为空!"
		}

		if row.OsName == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "操作系统不能为空!"
		}

		if row.SystemName == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "系统安装模板不能为空!"
		}

		if row.Location == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "位置不能为空!"
		}

		countDevice, err := repo.CountDeviceBySn(row.Sn)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		if countDevice > 0 {
			ID, err := repo.GetDeviceIdBySn(row.Sn)
			row.ID = ID
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}

			//hostname
			countHostname, err := repo.CountDeviceByHostnameAndId(row.Hostname, ID)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误:" + err.Error()})
				return
			}
			if countHostname > 0 {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "该主机名已存在!"
			}

			//IP
			countIp, err := repo.CountDeviceByIpAndId(row.Ip, ID)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}

			if countIp > 0 {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "该IP已存在!"
			}
		} else {
			//hostname
			countHostname, err := repo.CountDeviceByHostname(row.Hostname)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误:" + err.Error()})
				return
			}
			if countHostname > 0 {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "该主机名已存在!"
			}

			//IP
			countIp, err := repo.CountDeviceByIp(row.Ip)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}

			if countIp > 0 {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "该IP已存在!"
			}
		}

		//匹配网络
		isValidate, err := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", row.Ip)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
			return
		}

		if !isValidate {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "IP格式不正确!"
		}

		modelIp, err := repo.GetIpByIp(row.Ip)
		if err != nil {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "未匹配到网段!"
		} else {
			network, errNetwork := repo.GetNetworkById(modelIp.NetworkID)
			if errNetwork != nil {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "未匹配到网段!"
			}
			row.NetworkName = network.Network
		}

		//OSName
		countOs, err := repo.CountOsConfigByName(row.OsName)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		if countOs <= 0 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "未匹配到操作系统!"
		}

		//SystemName
		countSystem, err := repo.CountSystemConfigByName(row.SystemName)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		if countSystem <= 0 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "未匹配到系统安装模板!"
		}

		if row.HardwareName != "" {
			//HardwareName
			countHardware, err := repo.CountHardwarrWithSeparator(row.HardwareName)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}

			if countHardware <= 0 {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "未匹配到硬件配置模板!"
			}
		}

		/*
			if row.Location != "" {
				locationId, err := repo.GetLocationIdByName(row.Location)
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
					return
				}
				if locationId <= 0 {
					var br string
					if row.Content != "" {
						br = "<br />"
					}
					row.Content += br + "未匹配到位置!"
				} else {
					row.LocationID = locationId
				}
			}
		*/

		if row.Content != "" {
			failure = append(failure, row)
		} else {
			success = append(success, row)
		}
	}

	var data []Device
	if len(failure) > 0 {
		data = failure
	} else {
		data = success
	}
	var result []Device
	for i := 0; i < len(data); i++ {
		if uint(i) >= info.Offset && uint(i) < (info.Offset+info.Limit) {
			result = append(result, data[i])
		}
	}

	if len(failure) > 0 {
		w.WriteJSON(map[string]interface{}{"Status": "failure", "Message": "设备信息不正确", "recordCount": len(data), "Content": result})
	} else {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "recordCount": len(data), "Content": result})
	}
}

func ImportDevice(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		Filename string
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	file := "./upload/" + info.Filename

	cd, err := iconv.Open("utf-8", "gbk") // convert gbk to utf8
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}
	defer cd.Close()

	input, err := os.Open(file)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}
	bufSize := 1024 * 1024
	read := iconv.NewReader(cd, input, bufSize)

	r2 := csv.NewReader(read)
	ra, err := r2.ReadAll()
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	length := len(ra)

	type Device struct {
		ID              uint
		BatchNumber     string
		Sn              string
		Hostname        string
		Ip              string
		NetworkID       uint
		OsID            uint
		HardwareID      uint
		SystemID        uint
		Location        string
		LocationID      uint
		AssetNumber     string
		Status          string
		InstallProgress float64
		InstallLog      string
		NetworkName     string
		OsName          string
		HardwareName    string
		SystemName      string
		Content         string
		IsSupportVm     string
	}

	batchNumber, err := repo.CreateBatchNumber()
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	//var result []string
	for i := 1; i < length; i++ {
		//result = append(result, ra[i][0])
		var row Device

		if len(ra[i]) != 8 {
			continue
		}

		row.Sn = strings.TrimSpace(ra[i][0])
		row.Hostname = strings.TrimSpace(ra[i][1])
		row.Ip = strings.TrimSpace(ra[i][2])
		row.OsName = strings.TrimSpace(ra[i][3])
		row.HardwareName = strings.TrimSpace(ra[i][4])
		row.SystemName = strings.TrimSpace(ra[i][5])
		row.Location = strings.TrimSpace(ra[i][6])
		row.AssetNumber = strings.TrimSpace(ra[i][7])

		if len(row.Sn) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "SN长度超过255限制!"
		}

		if len(row.Hostname) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "主机名长度超过255限制!"
		}

		if len(row.Location) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "位置长度超过255限制!"
		}

		if len(row.AssetNumber) > 255 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "财编长度超过255限制!"
		}

		if row.Sn == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "SN不能为空!"
		}

		if row.Hostname == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "主机名不能为空!"
		}

		if row.Ip == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "IP不能为空!"
		}

		if row.OsName == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "操作系统不能为空!"
		}

		if row.SystemName == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "系统安装模板不能为空!"
		}

		if row.Location == "" {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "位置不能为空!"
		}

		countDevice, err := repo.CountDeviceBySn(row.Sn)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		if countDevice > 0 {
			ID, err := repo.GetDeviceIdBySn(row.Sn)
			row.ID = ID
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}

			//hostname
			countHostname, err := repo.CountDeviceByHostnameAndId(row.Hostname, ID)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误:" + err.Error()})
				return
			}
			if countHostname > 0 {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "SN:" + row.Sn + "该主机名已存在!"
			}

			//IP
			countIp, err := repo.CountDeviceByIpAndId(row.Ip, ID)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}

			if countIp > 0 {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "SN:" + row.Sn + "该IP已存在!"
			}
		} else {
			//hostname
			countHostname, err := repo.CountDeviceByHostname(row.Hostname)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误:" + err.Error()})
				return
			}
			if countHostname > 0 {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "SN:" + row.Sn + "该主机名已存在!"
			}

			//IP
			countIp, err := repo.CountDeviceByIp(row.Ip)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}

			if countIp > 0 {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "SN:" + row.Sn + "该IP已存在!"
			}
		}

		//匹配网络
		isValidate, err := regexp.MatchString("^((2[0-4]\\d|25[0-5]|[01]?\\d\\d?)\\.){3}(2[0-4]\\d|25[0-5]|[01]?\\d\\d?)$", row.Ip)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
			return
		}

		if !isValidate {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "SN:" + row.Sn + "IP格式不正确!"
		}

		modelIp, err := repo.GetIpByIp(row.Ip)
		if err != nil {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "SN:" + row.Sn + "未匹配到网段!"
		}

		_, errNetwork := repo.GetNetworkById(modelIp.NetworkID)
		if errNetwork != nil {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "SN:" + row.Sn + "未匹配到网段!"
		}

		row.NetworkID = modelIp.NetworkID

		//OSName
		countOs, err := repo.CountOsConfigByName(row.OsName)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		if countOs <= 0 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "SN:" + row.Sn + "未匹配到操作系统!"
		}
		mod, err := repo.GetOsConfigByName(row.OsName)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		row.OsID = mod.ID

		//SystemName
		countSystem, err := repo.CountSystemConfigByName(row.SystemName)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}

		if countSystem <= 0 {
			var br string
			if row.Content != "" {
				br = "<br />"
			}
			row.Content += br + "SN:" + row.Sn + "未匹配到系统安装模板!"
		}

		systemId, err := repo.GetSystemConfigIdByName(row.SystemName)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		row.SystemID = systemId

		if row.HardwareName != "" {
			//HardwareName
			countHardware, err := repo.CountHardwarrWithSeparator(row.HardwareName)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}

			if countHardware <= 0 {
				var br string
				if row.Content != "" {
					br = "<br />"
				}
				row.Content += br + "SN:" + row.Sn + "未匹配到硬件配置模板!"
			}

			hardware, err := repo.GetHardwareBySeaprator(row.HardwareName)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
			row.HardwareID = hardware.ID
		}

		if row.Location != "" {
			locationId, err := repo.GetLocationIdByName(row.Location)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
			if locationId <= 0 {
				/*
					var br string
					if row.Content != "" {
						br = "<br />"
					}
					row.Content += br + "SN:" + row.Sn + " 未匹配到位置!"
				*/
				_, err := repo.ImportLocation(row.Location)
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
					return
				}
				locationId, err := repo.GetLocationIdByName(row.Location)
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
					return
				}
				if locationId <= 0 {
					var br string
					if row.Content != "" {
						br = "<br />"
					}
					row.Content += br + "SN:" + row.Sn + " 未匹配到位置!"
				}
				row.LocationID = locationId
			} else {
				row.LocationID = locationId
			}
		}
		if row.Content != "" {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": row.Content})
			return
		} else {
			status := "pre_install"
			row.IsSupportVm = "Yes"
			if countDevice > 0 {
				id, err := repo.GetDeviceIdBySn(row.Sn)
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
					return
				}

				_, errUpdate := repo.UpdateDeviceById(id, batchNumber, row.Sn, row.Hostname, row.Ip, row.NetworkID, row.OsID, row.HardwareID, row.SystemID, "", row.LocationID, row.AssetNumber, status, row.IsSupportVm)
				if errUpdate != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "操作失败:" + errUpdate.Error()})
					return
				}
			} else {
				_, err := repo.AddDevice(batchNumber, row.Sn, row.Hostname, row.Ip, row.NetworkID, row.OsID, row.HardwareID, row.SystemID, "", row.LocationID, row.AssetNumber, status, row.IsSupportVm)
				if err != nil {
					w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "操作失败:" + err.Error()})
					return
				}
			}
		}
	}

	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
}