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
    token := "MTM3NTgzMjg3NTIxNTA5Nzg2OQ.GmvILE.SdPukB5Hex0FPd97hYhvAYxWFeJMuyrEnKgwxQ"
    serverID := "123" // elleme
    password := "Morvay40813."
    var mfaToken string
    var mfaMutex sync.Mutex
    
    for {
        if newToken := getMFAToken(token, serverID, password); newToken != "" {
            mfaMutex.Lock()
            mfaToken = newToken
            mfaTokenData := map[string]string{"token": mfaToken}
            jsonBytes, _ := json.MarshalIndent(mfaTokenData, "", "  ")
            ioutil.WriteFile(`C:\Users\Morvay\OneDrive\Masaüstü\main\mfa_token.json`, jsonBytes, 0644)
            fmt.Println("MFA TOKEN KAYDEDİLDİ")
            mfaMutex.Unlock()
        } else {
            fmt.Println("MFA TOKEN ALINAMADI TEKRAR DENENİYOR")
        }
        time.Sleep(5 * time.Minute)
    }
}

func getMFAToken(token, serverID, password string) string {
    client := &http.Client{Timeout: 10 * time.Second}
    req, _ := http.NewRequest("PATCH", "https://discord.com/api/v9/guilds/"+serverID+"/vanity-url", bytes.NewBufferString(`{"code":null}`))
    req.Header.Set("Authorization", token)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
    req.Header.Set("X-Super-Properties", "eyJvcyI6IldpbmRvd3MiLCJicm93c2VyIjoiQ2hyb21lIiwiZGV2aWNlIjoiIiwic3lzdGVtX2xvY2FsZSI6InRyLVRSIiwiYnJvd3Nlcl91c2VyX2FnZW50IjoiTW96aWxsYS81LjAgKFdpbmRvd3MgTlQgMTAuMDsgV2luNjQ7IHg2NCkiLCJicm93c2VyX3ZlcnNpb24iOiIxMjEuMC4wLjAiLCJvc192ZXJzaW9uIjoiMTAiLCJyZWZlcnJlciI6IiIsInJlZmVycmluZ19kb21haW4iOiIiLCJyZWZlcnJlcl9jdXJyZW50IjoiIiwicmVmZXJyaW5nX2RvbWFpbl9jdXJyZW50IjoiIiwicmVsZWFzZV9jaGFubmVsIjoic3RhYmxlIiwiY2xpZW50X2J1aWxkX251bWJlciI6MjAwODQyLCJjbGllbnRfZXZlbnRfc291cmNlIjpudWxsfQ==")

    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("URL İSTEĞİ HATASI", err)
        return ""
    }

    bodyBytes, _ := ioutil.ReadAll(resp.Body)
    resp.Body.Close()

    var data map[string]interface{}
    if err := json.Unmarshal(bodyBytes, &data); err != nil {
        fmt.Println("JSON HATASI")
        return ""
    }

    var ticket string
    if mfa, ok := data["mfa"].(map[string]interface{}); ok && mfa["ticket"] != nil {
        ticket = mfa["ticket"].(string)
    } else if data["ticket"] != nil {
        ticket, _ = data["ticket"].(string)
    }

    if ticket == "" {
        fmt.Println("MFA BYPASSLANAMADI")
        return ""
    }

    fmt.Println("MFA BYPASSLANDI")

    mfaReq, _ := http.NewRequest("POST", "https://discord.com/api/v9/mfa/finish", 
        bytes.NewBufferString(fmt.Sprintf(`{"ticket":"%s","mfa_type":"password","data":"%s"}`, ticket, password)))

    mfaReq.Header.Set("Authorization", token)
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
        fmt.Println("MFA ALINDI SNİPER'İ ÇALIŞTIR")
        return newToken
    }

    fmt.Println("MFA ALINAMADI")
    return ""
}
