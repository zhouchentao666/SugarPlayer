package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// 在线音乐评论：网易云 / QQ / 酷狗 / 酷我 / 咪咕 五个平台的歌曲评论抓取。
// 接口与字段映射参考 CeruMusic-main（src/main/utils/musicSdk/*/comment.js）。
// 所有平台均无需登录即可获取评论（游客态）。
// ---------------------------------------------------------------------------

// OnlineComment 是一条归一化后的评论（含楼中楼回复）。
type OnlineComment struct {
	ID         string          `json:"id"`
	Text       string          `json:"text"`
	Time       int64           `json:"time"` // 毫秒时间戳
	UserName   string          `json:"userName"`
	Avatar     string          `json:"avatar"`
	UserID     string          `json:"userId"`
	LikedCount int             `json:"likedCount"`
	Location   string          `json:"location"`
	Images     []string        `json:"images"`
	ReplyNum   int             `json:"replyNum"`
	Reply      []OnlineComment `json:"reply"`
}

// OnlineCommentPage 是一次评论拉取的结果页。
type OnlineCommentPage struct {
	Source string          `json:"source"`
	Kind   string          `json:"kind"`
	Comments []OnlineComment `json:"comments"`
	Total  int             `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
	MaxPage int            `json:"maxPage"`
}

var commentHTTPClient = &http.Client{Timeout: 20 * time.Second}

// OnlineComments 根据歌曲来源拉取评论。kind 为 "latest"（最新）或 "hot"（热门）。
func (a *App) OnlineComments(song OnlineSong, kind string, page int) (OnlineCommentPage, error) {
	if page < 1 {
		page = 1
	}
	if kind != "hot" {
		kind = "latest"
	}
	limit := 20
	switch song.Source {
	case "netease", "wy":
		return a.commentsWy(song.ID, kind, page, limit)
	case "qq", "tx":
		return a.commentsTx(song, kind, page, limit)
	case "kugou", "kg":
		return a.commentsKg(song.ID, kind, page, limit)
	case "kuwo", "kw":
		// go-music-dl 的 kuwo ID 已去掉 "MUSIC_" 前缀，评论接口需要完整 rid。
		rid := song.ID
		if !strings.HasPrefix(rid, "MUSIC_") {
			rid = "MUSIC_" + rid
		}
		return a.commentsKw(rid, kind, page, limit)
	case "migu", "mg":
		// go-music-dl 的 migu ID 形如 "contentId|resourceType|formatType"，评论接口只需 contentId。
		contentID := song.ID
		if idx := strings.Index(contentID, "|"); idx >= 0 {
			contentID = contentID[:idx]
		}
		return a.commentsMg(contentID, kind, page, limit)
	default:
		return OnlineCommentPage{}, fmt.Errorf("unsupported source: %s", song.Source)
	}
}

// proxyImg 把远程图片地址转成本地 /cover 代理，避免 WebView 跨域/防盗链问题。
func (a *App) proxyImg(raw string) string {
	if raw == "" {
		return ""
	}
	if !strings.HasPrefix(raw, "http") {
		return raw
	}
	return fmt.Sprintf("http://127.0.0.1:%d/cover?url=%s", a.audio.port, url.QueryEscape(raw))
}

// ---------------------------------------------------------------------------
// 通用 HTTP 辅助
// ---------------------------------------------------------------------------

func httpGetJSON(endpoint string, headers map[string]string, out interface{}) (int, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := commentHTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if out != nil {
		_ = json.Unmarshal(body, out)
	}
	return resp.StatusCode, nil
}

func httpPostForm(endpoint string, form url.Values, headers map[string]string, out interface{}) (int, error) {
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := commentHTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if out != nil {
		_ = json.Unmarshal(body, out)
	}
	return resp.StatusCode, nil
}

func httpPostJSON(endpoint string, payload interface{}, headers map[string]string, out interface{}) (int, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := commentHTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if out != nil {
		_ = json.Unmarshal(body, out)
	}
	return resp.StatusCode, nil
}

// ---------------------------------------------------------------------------
// 网易云（wy）：weapi 加密（AES-128-CBC 双重 + RSA 无填充）
// ---------------------------------------------------------------------------

const (
	neteaseIV      = "0102030405060708"
	neteasePreset  = "0CoJUm6Qyw8W8jud"
	neteaseBase62  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	neteasePubPEM  = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDgtQn2JZ34ZC28NWYpAUd98iZ37BUrX/aKzmFbt7clFSs6sXqHauqKWqdtLkF2KexO40H1YTX8z2lSgBBOAxLsvaklV8k4cBFK9snQXE9/DDaFt6Rr7iVZMldczhC0JNgTz+SHXT6CBHuX3e9SdB1Ua44oncaTWz7OBGLbCiK45wIDAQAB
-----END PUBLIC KEY-----`
)

var neteasePub *rsa.PublicKey

func init() {
	block, _ := pem.Decode([]byte(neteasePubPEM))
	if block != nil {
		if pub, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
			if rsaPub, ok := pub.(*rsa.PublicKey); ok {
				neteasePub = rsaPub
			}
		}
	}
}

func aesCBCEncrypt(plaintext, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	pad := aes.BlockSize - len(plaintext)%aes.BlockSize
	plaintext = append(plaintext, bytes.Repeat([]byte{byte(pad)}, pad)...)
	ciphertext := make([]byte, len(plaintext))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, plaintext)
	return ciphertext, nil
}

// rsaNoPaddingEncrypt 实现 Node 的 publicEncrypt({padding: RSA_NO_PADDING}, buffer)：
// 把 16 字节密钥右对齐补零到 128 字节后做原始 RSA 加密（无填充），结果转 256 位十六进制。
func rsaNoPaddingEncrypt(secretKey []byte) (string, error) {
	if neteasePub == nil {
		return "", fmt.Errorf("netease public key not loaded")
	}
	reversed := make([]byte, len(secretKey))
	for i, b := range secretKey {
		reversed[len(secretKey)-1-i] = b
	}
	buf := make([]byte, 128)
	copy(buf[128-len(reversed):], reversed)
	m := new(big.Int).SetBytes(buf)
	c := new(big.Int).Exp(m, big.NewInt(int64(neteasePub.E)), neteasePub.N)
	hexStr := hex.EncodeToString(c.Bytes())
	if len(hexStr) < 256 {
		hexStr = strings.Repeat("0", 256-len(hexStr)) + hexStr
	}
	return hexStr, nil
}

func weapi(object map[string]interface{}) (map[string]string, error) {
	text, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}
	inner, err := aesCBCEncrypt(text, []byte(neteasePreset), []byte(neteaseIV))
	if err != nil {
		return nil, err
	}
	innerB64 := base64.StdEncoding.EncodeToString(inner)

	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return nil, err
	}
	secKey := make([]byte, 16)
	for i, b := range randBytes {
		secKey[i] = neteaseBase62[int(b)%62]
	}

	outer, err := aesCBCEncrypt([]byte(innerB64), secKey, []byte(neteaseIV))
	if err != nil {
		return nil, err
	}
	params := base64.StdEncoding.EncodeToString(outer)
	encSecKey, err := rsaNoPaddingEncrypt(secKey)
	if err != nil {
		return nil, err
	}
	return map[string]string{"params": params, "encSecKey": encSecKey}, nil
}

type wyComment struct {
	CommentId   int64 `json:"commentId"`
	Content     string `json:"content"`
	Time        int64 `json:"time"`
	LikedCount  int    `json:"likedCount"`
	IpLocation  struct {
		Location string `json:"location"`
	} `json:"ipLocation"`
	User struct {
		Nickname  string `json:"nickname"`
		AvatarUrl string `json:"avatarUrl"`
		UserId    int64  `json:"userId"`
	} `json:"user"`
	BeReplied []struct {
		BeRepliedCommentId int64 `json:"beRepliedCommentId"`
		Content            string `json:"content"`
		User               struct {
			Nickname  string `json:"nickname"`
			AvatarUrl string `json:"avatarUrl"`
			UserId    int64  `json:"userId"`
		} `json:"user"`
		IpLocation struct {
			Location string `json:"location"`
		} `json:"ipLocation"`
	} `json:"beReplied"`
}

func (a *App) commentsWy(id, kind string, page, limit int) (OnlineCommentPage, error) {
	rid := "R_SO_4_" + id
	var endpoint string
	var obj map[string]interface{}
	if kind == "hot" {
		endpoint = "https://music.163.com/weapi/v1/resource/hotcomments/" + rid
		obj = map[string]interface{}{
			"rid":        rid,
			"limit":      limit,
			"offset":     limit * (page - 1),
			"beforeTime": strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10),
		}
	} else {
		endpoint = "https://music.163.com/weapi/comment/resource/comments/get"
		obj = map[string]interface{}{
			"cursor":    time.Now().UnixNano() / int64(time.Millisecond),
			"offset":    limit * (page - 1),
			"orderType": 1,
			"pageNo":    page,
			"pageSize":  limit,
			"rid":       rid,
			"threadId":  rid,
		}
	}
	form, err := weapi(obj)
	if err != nil {
		return OnlineCommentPage{}, err
	}
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.90 Safari/537.36",
		"origin":     "https://music.163.com",
		"Referer":    "http://music.163.com/",
	}
	var out struct {
		Code int `json:"code"`
		Data struct {
			Comments   []wyComment `json:"comments"`
			TotalCount int         `json:"totalCount"`
			Cursor     int64       `json:"cursor"`
		} `json:"data"`
		HotComments []wyComment `json:"hotComments"`
		Total       int         `json:"total"`
	}
	status, err := httpPostForm(endpoint, url.Values{
		"params":    {form["params"]},
		"encSecKey": {form["encSecKey"]},
	}, headers, &out)
	if err != nil {
		return OnlineCommentPage{}, err
	}
	if status != 200 || out.Code != 200 {
		return OnlineCommentPage{}, fmt.Errorf("网易云评论获取失败 (code %d)", out.Code)
	}
	raw := out.Data.Comments
	total := out.Data.TotalCount
	if kind == "hot" {
		raw = out.HotComments
		total = out.Total
	}
	comments := make([]OnlineComment, 0, len(raw))
	for _, c := range raw {
		comments = append(comments, a.filterWyComment(c))
	}
	return OnlineCommentPage{
		Source:   "wy",
		Kind:     kind,
		Comments: comments,
		Total:    total,
		Page:     page,
		Limit:    limit,
		MaxPage:  (total + limit - 1) / limit,
	}, nil
}

func (a *App) filterWyComment(c wyComment) OnlineComment {
	// 若有被回复对象：主评论显示「被回复的内容」，原评论者放入 reply（与 CeruMusic 一致）。
	if len(c.BeReplied) > 0 {
		r := c.BeReplied[0]
		original := OnlineComment{
			ID:         strconv.FormatInt(c.CommentId, 10),
			Text:       replaceWyEmoji(c.Content),
			Time:       c.Time,
			UserName:   c.User.Nickname,
			Avatar:     a.proxyImg(c.User.AvatarUrl),
			UserID:     strconv.FormatInt(c.User.UserId, 10),
			LikedCount: c.LikedCount,
			Location:   c.IpLocation.Location,
		}
		return OnlineComment{
			ID:         strconv.FormatInt(c.CommentId, 10),
			Text:       replaceWyEmoji(r.Content),
			Time:       c.Time,
			UserName:   r.User.Nickname,
			Avatar:     a.proxyImg(r.User.AvatarUrl),
			UserID:     strconv.FormatInt(r.User.UserId, 10),
			Location:   r.IpLocation.Location,
			Reply:      []OnlineComment{original},
		}
	}
	return OnlineComment{
		ID:         strconv.FormatInt(c.CommentId, 10),
		Text:       replaceWyEmoji(c.Content),
		Time:       c.Time,
		UserName:   c.User.Nickname,
		Avatar:     a.proxyImg(c.User.AvatarUrl),
		UserID:     strconv.FormatInt(c.User.UserId, 10),
		LikedCount: c.LikedCount,
		Location:   c.IpLocation.Location,
	}
}

// replaceWyEmoji 把网易云评论里的 [大笑] 等表情标记替换为对应 emoji。
func replaceWyEmoji(text string) string {
	if text == "" {
		return ""
	}
	for k, v := range wyEmojiMap {
		text = strings.ReplaceAll(text, "["+k+"]", v)
	}
	return text
}

var wyEmojiMap = map[string]string{
	"大笑": "😃", "可爱": "😊", "憨笑": "☺️", "色": "😍", "亲亲": "😙", "惊恐": "😱",
	"流泪": "😭", "亲": "😚", "呆": "😳", "哀伤": "😔", "呲牙": "😁", "吐舌": "😝",
	"撇嘴": "😒", "怒": "😡", "奸笑": "😏", "汗": "😓", "痛苦": "😖", "惶恐": "😰",
	"生病": "😨", "口罩": "😷", "大哭": "😂", "晕": "😵", "发怒": "👿", "开心": "😄",
	"鬼脸": "😜", "皱眉": "😞", "流感": "😢", "爱心": "❤️", "心碎": "💔", "钟情": "💘",
	"星星": "⭐️", "生气": "💢", "便便": "💩", "强": "👍", "弱": "👎", "拜": "🙏",
	"牵手": "👫", "跳舞": "👯‍♀️", "禁止": "🙅‍♀️", "这边": "💁‍♀️", "爱意": "💏",
	"示爱": "👩‍❤️‍👨", "嘴唇": "👄", "狗": "🐶", "猫": "🐱", "猪": "🐷", "兔子": "🐰",
	"小鸡": "🐤", "公鸡": "🐔", "幽灵": "👻", "圣诞": "🎅", "外星": "👽", "钻石": "💎",
	"礼物": "🎁", "男孩": "👦", "女孩": "👧", "蛋糕": "🎂", "18": "🔞", "圈": "⭕", "叉": "❌",
}

// ---------------------------------------------------------------------------
// QQ 音乐（tx）：需先把 songmid 解析为 songId
// ---------------------------------------------------------------------------

func (a *App) resolveTxSongId(songmid string) (string, error) {
	payload := map[string]interface{}{
		"comm": map[string]interface{}{"ct": "19", "cv": "1859", "uin": "0"},
		"req": map[string]interface{}{
			"module": "music.pf_song_detail_svr",
			"method": "get_song_detail_yqq",
			"param": map[string]interface{}{"song_type": 0, "song_mid": songmid},
		},
	}
	var out struct {
		Code int `json:"code"`
		Req  struct {
			Code int `json:"code"`
		Data struct {
			TrackInfo struct {
				Id   int64 `json:"id"`
				Mid  string `json:"mid"`
				File struct {
					MediaMid string `json:"media_mid"`
				} `json:"file"`
			} `json:"track_info"`
		} `json:"data"`
	} `json:"req"`
}
headers := map[string]string{
	"User-Agent": "Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0)",
}
status, err := httpPostJSON("https://u.y.qq.com/cgi-bin/musicu.fcg", payload, headers, &out)
if err != nil {
	return "", err
}
if status != 200 || out.Code != 0 || out.Req.Code != 0 || out.Req.Data.TrackInfo.Id == 0 {
	return "", fmt.Errorf("QQ 歌曲 ID 解析失败")
}
return strconv.FormatInt(out.Req.Data.TrackInfo.Id, 10), nil
}

type txCommentItem struct {
	RootCommentId   string `json:"rootcommentid"`
	CommentId       string `json:"commentid"`
	RootContent     string `json:"rootcommentcontent"`
	RootNick        string `json:"rootcommentnick"`
	Avatar          string `json:"avatarurl"`
	EncryptRootUin  string `json:"encrypt_rootcommentuin"`
	PraiseNum       int    `json:"praisenum"`
	Time            int64  `json:"time"`
	Middle          []struct {
		SubCommentId string `json:"subcommentid"`
		Content      string `json:"subcommentcontent"`
		ReplyNick    string `json:"replynick"`
		Avatar       string `json:"avatarurl"`
		EncryptUin   string `json:"encrypt_replyuin"`
		PraiseNum    int    `json:"praisenum"`
		Time         int64  `json:"time"`
	} `json:"middlecommentcontent"`
}

func (a *App) commentsTx(song OnlineSong, kind string, page, limit int) (OnlineCommentPage, error) {
	songId, err := a.resolveTxSongId(song.ID)
	if err != nil {
		return OnlineCommentPage{}, err
	}
	if kind == "hot" {
		return a.commentsTxHot(songId, page, limit)
	}
	form := url.Values{}
	form.Set("uin", "0")
	form.Set("format", "json")
	form.Set("cid", "205360772")
	form.Set("reqtype", "2")
	form.Set("biztype", "1")
	form.Set("topid", songId)
	form.Set("cmd", "8")
	form.Set("needmusiccrit", "1")
	form.Set("pagenum", strconv.Itoa(page-1))
	form.Set("pagesize", strconv.Itoa(limit))
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0)",
	}
	var out struct {
		Code    int `json:"code"`
		Comment struct {
			CommentList  []txCommentItem `json:"commentlist"`
			CommentTotal int             `json:"commenttotal"`
		} `json:"comment"`
	}
	status, err := httpPostForm("http://c.y.qq.com/base/fcgi-bin/fcg_global_comment_h5.fcg", form, headers, &out)
	if err != nil {
		return OnlineCommentPage{}, err
	}
	if status != 200 || out.Code != 0 {
		return OnlineCommentPage{}, fmt.Errorf("QQ 评论获取失败 (code %d)", out.Code)
	}
	comments := make([]OnlineComment, 0, len(out.Comment.CommentList))
	for _, c := range out.Comment.CommentList {
		comments = append(comments, a.filterTxNew(c))
	}
	total := out.Comment.CommentTotal
	return OnlineCommentPage{
		Source:   "tx",
		Kind:     kind,
		Comments: comments,
		Total:    total,
		Page:     page,
		Limit:    limit,
		MaxPage:  (total + limit - 1) / limit,
	}, nil
}

func (a *App) filterTxNew(c txCommentItem) OnlineComment {
	oc := OnlineComment{
		ID:         c.RootCommentId + "_" + c.CommentId,
		Text:       replaceTxEmoji(strings.ReplaceAll(c.RootContent, "\\n", "\n")),
		Time:       txNormTime(c.Time),
		UserName:   strings.TrimPrefix(c.RootNick, "*"),
		Avatar:     a.proxyImg(c.Avatar),
		UserID:     c.EncryptRootUin,
		LikedCount: c.PraiseNum,
	}
	for _, m := range c.Middle {
		oc.Reply = append(oc.Reply, OnlineComment{
			ID:         "sub_" + c.RootCommentId + "_" + m.SubCommentId,
			Text:       replaceTxEmoji(strings.ReplaceAll(m.Content, "\\n", "\n")),
			Time:       txNormTime(m.Time),
			UserName:   strings.TrimPrefix(m.ReplyNick, "*"),
			Avatar:     a.proxyImg(m.Avatar),
			UserID:     m.EncryptUin,
			LikedCount: m.PraiseNum,
		})
	}
	return oc
}

func (a *App) commentsTxHot(songId string, page, limit int) (OnlineCommentPage, error) {
	payload := map[string]interface{}{
		"comm": map[string]interface{}{
			"cv": 4747474, "ct": 24, "format": "json", "inCharset": "utf-8",
			"outCharset": "utf-8", "notice": 0, "platform": "yqq.json",
			"needNewCode": 1, "uin": 0,
		},
		"req": map[string]interface{}{
			"module": "music.globalComment.CommentRead",
			"method": "GetHotCommentList",
			"param": map[string]interface{}{
				"BizType": 1, "BizId": songId, "LastCommentSeqNo": "",
				"PageSize": limit, "PageNum": page - 1, "HotType": 1,
				"WithAirborne": 0, "PicEnable": 1,
			},
		},
	}
	var out struct {
		Code int `json:"code"`
		Req  struct {
			Code int `json:"code"`
			Data struct {
				CommentList struct {
					Comments []struct {
						SeqNo    string `json:"SeqNo"`
						CmId     string `json:"CmId"`
						Content  string `json:"Content"`
						PubTime  int64  `json:"PubTime"`
						Nick     string `json:"Nick"`
						Avatar   string `json:"Avatar"`
						Location string `json:"Location"`
						EncryptUin string `json:"EncryptUin"`
						PraiseNum  int    `json:"PraiseNum"`
						Pic       string `json:"Pic"`
						SubComments []struct {
							SeqNo      string `json:"SeqNo"`
							CmId       string `json:"CmId"`
							Content    string `json:"Content"`
							PubTime    int64  `json:"PubTime"`
							Nick       string `json:"Nick"`
							Avatar     string `json:"Avatar"`
							EncryptUin string `json:"EncryptUin"`
							PraiseNum  int    `json:"PraiseNum"`
							Pic        string `json:"Pic"`
						} `json:"SubComments"`
					} `json:"Comments"`
					Total int `json:"Total"`
				} `json:"CommentList"`
			} `json:"data"`
		} `json:"req"`
	}
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36 Edg/113.0.0.0",
		"referer":    "https://y.qq.com/",
		"origin":     "https://y.qq.com",
	}
	status, err := httpPostJSON("https://u.y.qq.com/cgi-bin/musicu.fcg", payload, headers, &out)
	if err != nil {
		return OnlineCommentPage{}, err
	}
	if status != 200 || out.Code != 0 || out.Req.Code != 0 {
		return OnlineCommentPage{}, fmt.Errorf("QQ 热门评论获取失败")
	}
	list := out.Req.Data.CommentList.Comments
	comments := make([]OnlineComment, 0, len(list))
	for _, c := range list {
		oc := OnlineComment{
			ID:         c.SeqNo + "_" + c.CmId,
			Text:       replaceTxEmoji(strings.ReplaceAll(c.Content, "\\n", "\n")),
			Time:       txNormTime(c.PubTime),
			UserName:   c.Nick,
			Avatar:     a.proxyImg(c.Avatar),
			Location:   c.Location,
			UserID:     c.EncryptUin,
			LikedCount: c.PraiseNum,
		}
		if c.Pic != "" {
			oc.Images = []string{a.proxyImg(c.Pic)}
		}
		for _, s := range c.SubComments {
			sub := OnlineComment{
				ID:         "sub_" + s.SeqNo + "_" + s.CmId,
				Text:       replaceTxEmoji(strings.ReplaceAll(s.Content, "\\n", "\n")),
				Time:       txNormTime(s.PubTime),
				UserName:   s.Nick,
				Avatar:     a.proxyImg(s.Avatar),
				UserID:     s.EncryptUin,
				LikedCount: s.PraiseNum,
			}
			if s.Pic != "" {
				sub.Images = []string{a.proxyImg(s.Pic)}
			}
			oc.Reply = append(oc.Reply, sub)
		}
		comments = append(comments, oc)
	}
	total := out.Req.Data.CommentList.Total
	return OnlineCommentPage{
		Source:   "tx",
		Kind:     "hot",
		Comments: comments,
		Total:    total,
		Page:     page,
		Limit:    limit,
		MaxPage:  (total + limit - 1) / limit,
	}, nil
}

func txNormTime(t int64) int64 {
	s := strconv.FormatInt(t, 10)
	if len(s) < 10 {
		return 0
	}
	return t * 1000
}

// replaceTxEmoji 把 [em]e400846[/em] 这类 QQ 表情标记替换为 emoji。
func replaceTxEmoji(msg string) string {
	if msg == "" {
		return ""
	}
	for code, emoji := range txEmojiMap {
		msg = strings.ReplaceAll(msg, "[em]"+code+"[/em]", emoji)
	}
	return msg
}

var txEmojiMap = map[string]string{
	"e400846": "😘", "e400874": "😴", "e400825": "😃", "e400847": "😙", "e400835": "😍",
	"e400873": "😳", "e400836": "😎", "e400867": "😭", "e400832": "😊", "e400837": "😏",
	"e400875": "😫", "e400831": "😉", "e400855": "😡", "e400823": "😄", "e400862": "😨",
	"e400844": "😖", "e400841": "😓", "e400830": "😈", "e400828": "😆", "e400833": "😋",
	"e400822": "😀", "e400843": "😕", "e400829": "😇", "e400824": "😂", "e400834": "😌",
	"e400877": "😷", "e400132": "🍉", "e400181": "🍺", "e401067": "☕️", "e400186": "🥧",
	"e400343": "🐷", "e400116": "🌹", "e400126": "🍃", "e400613": "💋", "e401236": "❤️",
	"e400622": "💔", "e400637": "💣", "e400643": "💩", "e400773": "🔪", "e400102": "🌛",
	"e401328": "🌞", "e400420": "👏", "e400914": "🙌", "e400408": "👍", "e400414": "👎",
	"e401121": "✋", "e400396": "👋", "e400384": "👉", "e401115": "✊", "e400402": "👌",
	"e400905": "🙈", "e400906": "🙉", "e400907": "🙊", "e400562": "👻", "e400932": "🙏",
	"e400644": "💪", "e400611": "💉", "e400185": "🎁", "e400655": "💰", "e400325": "🐥",
	"e400612": "💊", "e400198": "🎉", "e401685": "⚡️", "e400631": "💝", "e400768": "🔥",
	"e400432": "👑",
}

// ---------------------------------------------------------------------------
// 酷狗（kg）：参数 MD5 签名
// ---------------------------------------------------------------------------

const (
	kgCommentKey  = "OIlwieks28dk2k092lksi2UIkp"
	kgCommentMid  = "16249512204336365674023395779019"
	kgCommentCode = "fc4be23b4e972707f36b8a828a93ba8a"
)

func kgCommentSignature(params string) string {
	parts := strings.Split(params, "&")
	sort.Strings(parts)
	joined := strings.Join(parts, "")
	sum := md5.Sum([]byte(kgCommentKey + joined + kgCommentKey))
	return hex.EncodeToString(sum[:])
}

func (a *App) commentsKg(hash, kind string, page, limit int) (OnlineCommentPage, error) {
	ts := time.Now().UnixMilli()
	params := fmt.Sprintf(
		"dfid=0&mid=%s&clienttime=%d&uuid=0&extdata=%s&appid=1005&code=%s&schash=%s&clientver=11409&p=%d&clienttoken=&pagesize=%d&ver=10&kugouid=0",
		kgCommentMid, ts, hash, kgCommentCode, hash, page, limit,
	)
	path := "newest"
	if kind == "hot" {
		path = "topliked"
	}
	endpoint := fmt.Sprintf("http://m.comment.service.kugou.com/r/v1/rank/%s?%s&signature=%s", path, params, kgCommentSignature(params))
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36 Edg/107.0.1418.24",
	}
	var out struct {
		ErrCode int `json:"err_code"`
		Count   int `json:"count"`
		List    []struct {
			Id       string `json:"id"`
			Content  string `json:"content"`
			UserName string `json:"user_name"`
			UserPic  string `json:"user_pic"`
			UserId   string `json:"user_id"`
			AddTime  int64  `json:"addtime"`
			Location string `json:"location"`
			Like     struct {
				LikeNum int `json:"likenum"`
			} `json:"like"`
			ReplyNum int      `json:"reply_num"`
			Images   []struct {
				Url string `json:"url"`
			} `json:"images"`
			PContent string `json:"pcontent"`
			PUser    string `json:"puser"`
			PUserId  string `json:"puser_id"`
		} `json:"list"`
	}
	status, err := httpGetJSON(endpoint, headers, &out)
	if err != nil {
		return OnlineCommentPage{}, err
	}
	if status != 200 || out.ErrCode != 0 {
		return OnlineCommentPage{}, fmt.Errorf("酷狗评论获取失败 (err %d)", out.ErrCode)
	}
	comments := make([]OnlineComment, 0, len(out.List))
	for _, c := range out.List {
		imgs := make([]string, 0, len(c.Images))
		for _, im := range c.Images {
			imgs = append(imgs, a.proxyImg(im.Url))
		}
		original := OnlineComment{
			ID:         c.Id,
			Text:       c.Content,
			Time:       c.AddTime * 1000,
			UserName:   c.UserName,
			Avatar:     a.proxyImg(c.UserPic),
			UserID:     c.UserId,
			LikedCount: c.Like.LikeNum,
			Location:   c.Location,
			ReplyNum:   c.ReplyNum,
			Images:     imgs,
		}
		// 若有父评论：主评论显示父评论内容，原评论放入 reply（与 CeruMusic 一致）。
		if c.PContent != "" {
			comments = append(comments, OnlineComment{
				ID:       c.Id,
				Text:     c.PContent,
				UserName: c.PUser,
				UserID:   c.PUserId,
				Reply:    []OnlineComment{original},
			})
		} else {
			comments = append(comments, original)
		}
	}
	total := out.Count
	return OnlineCommentPage{
		Source:   "kg",
		Kind:     kind,
		Comments: comments,
		Total:    total,
		Page:     page,
		Limit:    limit,
		MaxPage:  (total + limit - 1) / limit,
	}, nil
}

// ---------------------------------------------------------------------------
// 酷我（kw）
// ---------------------------------------------------------------------------

func (a *App) commentsKw(rid, kind string, page, limit int) (OnlineCommentPage, error) {
	typ := "get_comment"
	if kind == "hot" {
		typ = "get_rec_comment"
	}
	endpoint := fmt.Sprintf(
		"http://ncomment.kuwo.cn/com.s?f=web&type=%s&aapiver=1&prod=kwplayer_ar_10.5.2.0&digest=15&sid=%s&start=%d&msgflag=1&count=%d&newver=3&uid=0",
		typ, rid, limit*(page-1), limit,
	)
	headers := map[string]string{
		"User-Agent": "Dalvik/2.1.0 (Linux; U; Android 9;)",
	}
	var out struct {
		Code            string `json:"code"`
		Comments        []kwComment `json:"comments"`
		HotComments     []kwComment `json:"hot_comments"`
		CommentsCounts  int        `json:"comments_counts"`
		HotCommentsCounts int      `json:"hot_comments_counts"`
	}
	status, err := httpGetJSON(endpoint, headers, &out)
	if err != nil {
		return OnlineCommentPage{}, err
	}
	if status != 200 || out.Code != "200" {
		return OnlineCommentPage{}, fmt.Errorf("酷我评论获取失败")
	}
	raw := out.Comments
	total := out.CommentsCounts
	if kind == "hot" {
		raw = out.HotComments
		total = out.HotCommentsCounts
	}
	comments := make([]OnlineComment, 0, len(raw))
	for _, c := range raw {
		comments = append(comments, a.filterKwComment(c))
	}
	return OnlineCommentPage{
		Source:   "kw",
		Kind:     kind,
		Comments: comments,
		Total:    total,
		Page:     page,
		Limit:    limit,
		MaxPage:  (total + limit - 1) / limit,
	}, nil
}

type kwComment struct {
	Id        string `json:"id"`
	Msg       string `json:"msg"`
	Time      int64  `json:"time"`
	UName     string `json:"u_name"`
	UPic      string `json:"u_pic"`
	UId       string `json:"u_id"`
	LikeNum   int    `json:"like_num"`
	Mpic      string `json:"mpic"`
	ChildComments []kwComment `json:"child_comments"`
}

func (a *App) filterKwComment(c kwComment) OnlineComment {
	oc := OnlineComment{
		ID:         c.Id,
		Text:       c.Msg,
		Time:       c.Time * 1000,
		UserName:   c.UName,
		Avatar:     a.proxyImg(c.UPic),
		UserID:     c.UId,
		LikedCount: c.LikeNum,
	}
	if c.Mpic != "" {
		oc.Images = []string{a.proxyImg(decodeWK(c.Mpic))}
	}
	for _, ch := range c.ChildComments {
		oc.Reply = append(oc.Reply, OnlineComment{
			ID:         ch.Id,
			Text:       ch.Msg,
			Time:       ch.Time * 1000,
			UserName:   ch.UName,
			Avatar:     a.proxyImg(ch.UPic),
			UserID:     ch.UId,
			LikedCount: ch.LikeNum,
			Images:     imgOrEmpty(a, ch.Mpic),
		})
	}
	return oc
}

func imgOrEmpty(a *App, mpic string) []string {
	if mpic == "" {
		return nil
	}
	return []string{a.proxyImg(decodeWK(mpic))}
}

func decodeWK(s string) string {
	if decoded, err := url.QueryUnescape(s); err == nil {
		return decoded
	}
	return s
}

// ---------------------------------------------------------------------------
// 咪咕（mg）
// ---------------------------------------------------------------------------

func (a *App) commentsMg(songId, kind string, page, limit int) (OnlineCommentPage, error) {
	var endpoint string
	if kind == "hot" {
		endpoint = fmt.Sprintf(
			"https://app.c.nf.migu.cn/MIGUM3.0/user/comment/stack/v1.0?pageSize=%d&queryType=2&resourceId=%s&resourceType=2&hotCommentStart=%d",
			limit, songId, (page-1)*limit,
		)
	} else {
		endpoint = fmt.Sprintf(
			"https://app.c.nf.migu.cn/MIGUM3.0/user/comment/stack/v1.0?pageSize=%d&queryType=1&resourceId=%s&resourceType=2&commentId=",
			limit, songId,
		)
	}
	headers := map[string]string{
		"User-Agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1",
	}
	var out struct {
		Code string `json:"code"`
		Data struct {
			Comments       []mgComment `json:"comments"`
			HotComments    []mgComment `json:"hotComments"`
			CommentNums    string     `json:"commentNums"`
			CfgHotCount    string     `json:"cfgHotCount"`
		} `json:"data"`
	}
	status, err := httpGetJSON(endpoint, headers, &out)
	if err != nil {
		return OnlineCommentPage{}, err
	}
	if status != 200 || out.Code != "000000" {
		return OnlineCommentPage{}, fmt.Errorf("咪咕评论获取失败")
	}
	raw := out.Data.Comments
	totalStr := out.Data.CommentNums
	if kind == "hot" {
		raw = out.Data.HotComments
		totalStr = out.Data.CfgHotCount
	}
	comments := make([]OnlineComment, 0, len(raw))
	for _, c := range raw {
		comments = append(comments, a.filterMgComment(c))
	}
	total, _ := strconv.Atoi(totalStr)
	return OnlineCommentPage{
		Source:   "mg",
		Kind:     kind,
		Comments: comments,
		Total:    total,
		Page:     page,
		Limit:    limit,
		MaxPage:  (total + limit - 1) / limit,
	}, nil
}

type mgComment struct {
	CommentId   string `json:"commentId"`
	CommentInfo string `json:"commentInfo"`
	CommentTime int64  `json:"commentTime"`
	User        struct {
		NickName  string `json:"nickName"`
		MiddleIcon string `json:"middleIcon"`
		BigIcon   string `json:"bigIcon"`
		SmallIcon string `json:"smallIcon"`
		UserId    string `json:"userId"`
	} `json:"user"`
	OpNumItem struct {
		ThumbNum int `json:"thumbNum"`
	} `json:"opNumItem"`
	ReplyTotalCount int `json:"replyTotalCount"`
	ReplyComments   []struct {
		ReplyId    string `json:"replyId"`
		ReplyInfo  string `json:"replyInfo"`
		ReplyTime  int64  `json:"replyTime"`
		User       struct {
			NickName   string `json:"nickName"`
			MiddleIcon string `json:"middleIcon"`
			BigIcon    string `json:"bigIcon"`
			SmallIcon  string `json:"smallIcon"`
			UserId     string `json:"userId"`
		} `json:"user"`
	} `json:"replyComments"`
}

func (a *App) filterMgComment(c mgComment) OnlineComment {
	avatar := c.User.MiddleIcon
	if avatar == "" {
		avatar = c.User.BigIcon
	}
	if avatar == "" {
		avatar = c.User.SmallIcon
	}
	oc := OnlineComment{
		ID:         c.CommentId,
		Text:       c.CommentInfo,
		Time:       c.CommentTime,
		UserName:   c.User.NickName,
		Avatar:     a.proxyImg(avatar),
		UserID:     c.User.UserId,
		LikedCount: c.OpNumItem.ThumbNum,
		ReplyNum:   c.ReplyTotalCount,
	}
	for _, r := range c.ReplyComments {
		rav := r.User.MiddleIcon
		if rav == "" {
			rav = r.User.BigIcon
		}
		if rav == "" {
			rav = r.User.SmallIcon
		}
		oc.Reply = append(oc.Reply, OnlineComment{
			ID:         r.ReplyId,
			Text:       r.ReplyInfo,
			Time:       r.ReplyTime,
			UserName:   r.User.NickName,
			Avatar:     a.proxyImg(rav),
			UserID:     r.User.UserId,
		})
	}
	return oc
}
