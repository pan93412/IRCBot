// GolangIRCBot: 使用 Socket 連結的 IRC 機器人。
//
// STATUS: DEVELOPING

package main

import (
    "errors"
    "fmt"
    "log"
    "crypto/tls"
    "strings"
    "io"
    //"io/ioutil"
    //"time"
)

// CONST: \0 符號
const NULLCHAR = `\0`

// IRC 機器人的相關設定
type IRCInfo struct {
    // 使用者設定
    Username string // 使用者實際的 IRC ID
    Realname string // 使用者的真實名稱
    Nick     string // 使用者的 IRC ID 暱稱 (Nickname)
    Password string // 使用者密碼，將用於 SASL 驗證。若無需執行 SASL 驗證請留空。
    
    // 伺服器設定
    Server  string // 伺服器位址，例如：irc.freenode.net
    Port    string // 伺服器埠號
    
    // SASL PLAIN 設定
    SASLEnabled bool // 是否啟用 SASL 驗證
    
    // 個人化設定
    ChannelToJoin []string // 要加入的 Channel。若無需則請傳入一個空陣列。
}

// 檢查 IRCInfo 內容是否有效
func (IRC IRCInfo) CheckVaild() error {
    var issues string
    
    switch {
        case IRC.Username == "":
            issues += "(X) IRC 的使用者名稱是空的！\n"
        case IRC.Realname == "":
            issues += "(X) IRC 的真實名稱是空的！\n"
        case IRC.Nick == "":
            issues += "(X) IRC 的暱稱是空的！若不想設定暱稱，請設定與 Username 相同值。\n"
        case IRC.Password == "" && IRC.SASLEnabled:
            issues += "(X) 設定了 SASL 驗證但密碼為空。\n"
        case IRC.Server == "" || IRC.Port == "0":
            issues += "(X) IRC 伺服器未設定或埠號為空值。\n"
    }
    
    if issues == "" {
        return nil
    } else {
        return errors.New(issues)
    }
}

// 透過 TLS/SSL 登入 IRC 機器人
//
// verbose 為 bool 值，會顯示發送的指令以及過程。
// 若不加則不顯示任何訊息，除非發生問題。
func (IRC *IRCInfo) ConnectTLS(verbose bool) (*tls.Conn, error) {
    // 建立相關需要參數
    ircserv := IRC.Server + ":" + IRC.Port
    data := make([]byte, 2000)
    
    // 開始建立連線
    if verbose {
        log.Printf("[I] 建立連線：連線到 %s (TLS 連線)", ircserv)
    }
    conn, errConn := tls.Dial("tcp", ircserv, nil)
    var remoteAddr = conn.RemoteAddr().String()
    
    if errConn != nil {
        if verbose {
            log.Printf("[E][失敗] 建立連線：連線到 %s 失敗。原因：%v", ircserv, errConn)
        }
        return nil, errConn
    }
    
    if verbose {
        log.Printf("[I] 連線 %s (%s) 成功：開始登入使用者", remoteAddr, conn.RemoteAddr().Network())
    }
    
    _, errCAPLS := io.WriteString(conn, "CAP LS")
    if errCAPLS != nil {panic(errCAPLS)}
    
    _, errNick := io.WriteString(conn, "NICK " + IRC.Nick)
    if errNick != nil {panic(errNick)}
    
    _, errUser := io.WriteString(conn, "USER %s 8 * :%s" + IRC.Nick)
    if errUser != nil {panic(errUser)}
    
    for {
        dataLen, _ := conn.Read(data)
        if dataLen == 0 {
            continue
        }
        if strings.Contains(string(data[:dataLen]), "No Ident response") {
            if verbose {
                log.Printf("[I] 登入成功！開始進行後續動作。")
            }
            break
        }
    }
    
    if IRC.SASLEnabled {
        if verbose {log.Printf("[I] 開始請求 SASL 能力。")
            // CAP REQ :sasl 請求 SASL 能力
            _, err := io.WriteString(conn, "CAP REQ SASL")
            if err != nil {panic(err)}
            
            if verbose {log.Printf("[I] 開始驗證程序")}
            // AUTHENTICATE PLAIN 開始以 SASL PLAIN 驗證
            io.WriteString(conn, "CAP REQ SASL")
            for {
                dataLen, _ := conn.Read(data)
                if dataLen == 0 {
                    continue
                }
                fmt.Println(string(data[:dataLen]))
                if strings.Contains(string(data[:dataLen]), "AUTHENTICATE +") {
                    if verbose {
                        log.Printf("[I] 伺服器允許驗證")
                    }
                } else {
                    log.Printf("[D] AUTHENTICATE +, isn't received.")
                    continue
                }
            }
        }
    }
    
    //CloseConnection:
    if verbose {
        log.Printf("[I] 關閉連線：關閉對 %s 的連線。", remoteAddr)
    }
    conn.Close() // 結束連線
    return nil, nil
}

func main() {
    var IRCD IRCInfo = IRCInfo {
        Username: "OAO",
        Realname: "OWO",
        Nick: "QAQ",
        Password: ":PWD:OAO:OWO",
        Server: "irc.freenode.net",
        Port: "7000",
        SASLEnabled: true,
        ChannelToJoin: nil,
    }
    
    fmt.Println(IRCD.ConnectTLS(true))
}
