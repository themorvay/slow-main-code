package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"
    "sync"
)

func main() {
    token := "tokeninizi girin"
    serverID := "123" // elleme
    password := "şifrenizi girin"
    var mfaToken string
    var mfaMutex sync.Mutex
    
    for {
        if newToken := a7b8c9(token, serverID, password); newToken != "" {
            mfaMutex.Lock()
            mfaToken = newToken
            mfaTokenData := map[string]string{"token": mfaToken}
            jsonBytes, _ := json.MarshalIndent(mfaTokenData, "", "  ")
            ioutil.WriteFile(`C:\Users\Morvay\OneDrive\Masaüstü\main\mfa_token.json`, jsonBytes, 0644)
            fmt.Println("mfa gecildi pampa")
            mfaMutex.Unlock()
        } else {
            fmt.Println("mfa token alınamadı tekrar deneniyor")
        }
        time.Sleep(5 * time.Minute)
    }
}

func a7b8c9(x, y, z string) string {
    client := &http.Client{Timeout: 10 * time.Second}
    req, _ := http.NewRequest("PATCH", "https://discord.com/api/v9/guilds/"+y+"/vanity-url", bytes.NewBufferString(`{"code":null}`))
    req.Header.Set("Authorization", x)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
    req.Header.Set("X-Super-Properties", "eyJvcyI6IldpbmRvd3MiLCJicm93c2VyIjoiQ2hyb21lIiwiZGV2aWNlIjoiIiwic3lzdGVtX2xvY2FsZSI6InRyLVRSIiwiYnJvd3Nlcl91c2VyX2FnZW50IjoiTW96aWxsYS81LjAgKFdpbmRvd3MgTlQgMTAuMDsgV2luNjQ7IHg2NCkiLCJicm93c2VyX3ZlcnNpb24iOiIxMjEuMC4wLjAiLCJvc192ZXJzaW9uIjoiMTAiLCJyZWZlcnJlciI6IiIsInJlZmVycmluZ19kb21haW4iOiIiLCJyZWZlcnJlcl9jdXJyZW50IjoiIiwicmVmZXJyaW5nX2RvbWFpbl9jdXJyZW50IjoiIiwicmVsZWFzZV9jaGFubmVsIjoic3RhYmxlIiwiY2xpZW50X2J1aWxkX251bWJlciI6MjAwODQyLCJjbGllbnRfZXZlbnRfc291cmNlIjpudWxsfQ==")

    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("istek gönderememe hatası", err)
        return ""
    }

    bodyBytes, _ := ioutil.ReadAll(resp.Body)
    resp.Body.Close()

    var data map[string]interface{}
    if err := json.Unmarshal(bodyBytes, &data); err != nil {
        fmt.Println("json hatası")
        return ""
    }

    var ticket string
    if mfa, ok := data["mfa"].(map[string]interface{}); ok && mfa["ticket"] != nil {
        ticket = mfa["ticket"].(string)
    } else if data["ticket"] != nil {
        ticket, _ = data["ticket"].(string)
    }

    if ticket == "" {
        fmt.Println("mfa oturum açılamadı")
        return ""
    }

    fmt.Println("mfa gecildi pampa")

    mfaReq, _ := http.NewRequest("POST", "https://discord.com/api/v9/mfa/finish", 
        bytes.NewBufferString(fmt.Sprintf(`{"ticket":"%s","mfa_type":"password","data":"%s"}`, ticket, z)))

    mfaReq.Header.Set("Authorization", x)
    mfaReq.Header.Set("Content-Type", "application/json")
    mfaReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
    mfaReq.Header.Set("X-Super-Properties", "eyJvcyI6IldpbmRvd3MiLCJicm93c2VyIjoiQ2hyb21lIiwiZGV2aWNlIjoiIiwic3lzdGVtX2xvY2FsZSI6InRyLVRSIiwiYnJvd3Nlcl91c2VyX2FnZW50IjoiTW96aWxsYS81LjAgKFdpbmRvd3MgTlQgMTAuMDsgV2luNjQ7IHg2NCkiLCJicm93c2VyX3ZlcnNpb24iOiIxMjEuMC4wLjAiLCJvc192ZXJzaW9uIjoiMTAiLCJyZWZlcnJlciI6IiIsInJlZmVycmluZ19kb21haW4iOiIiLCJyZWZlcnJlcl9jdXJyZW50IjoiIiwicmVmZXJyaW5nX2RvbWFpbl9jdXJyZW50IjoiIiwicmVsZWFzZV9jaGFubmVsIjoic3RhYmxlIiwiY2xpZW50X2J1aWxkX251bWJlciI6MjAwODQyLCJjbGllbnRfZXZlbnRfc291cmNlIjpudWxsfQ==")

    mfaResp, err := client.Do(mfaReq)
    if err != nil {
        fmt.Println("İSTEK YOLLANAMADI", err)
        return ""
    }

    mfaBytes, _ := ioutil.ReadAll(mfaResp.Body)
    mfaResp.Body.Close()

    var tokenData map[string]interface{}
    if json.Unmarshal(mfaBytes, &tokenData) != nil {
        fmt.Println("JSON PARSING HATASI")
        return ""
    }

    if newToken, ok := tokenData["token"].(string); ok {
        fmt.Println("mfa alındı sniperini çalıştır")
        return newToken
    }

    fmt.Println("mfa alamadım")
    return ""
}

func rev(s string) string {
    r := []rune(s)
    for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
        r[i], r[j] = r[j], r[i]
    }
    return string(r)
}
