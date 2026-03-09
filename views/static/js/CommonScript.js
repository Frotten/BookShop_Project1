const ACCESS_TOKEN_KEY = "access_token";
const CODE_SUCCESS = 1000;
const CODE_INVALID_TOKEN = 1007;
const CODE_NEED_LOGIN = 1008;

console.log("CommonScript loaded");

function getAccessToken() {
    return localStorage.getItem(ACCESS_TOKEN_KEY);
}

function setAccessToken(token) {
    if (token) {
        localStorage.setItem(ACCESS_TOKEN_KEY, token);
    }
}

function clearAccessToken() {
    localStorage.removeItem(ACCESS_TOKEN_KEY);
}

// 解析 JWT payload（不验证签名，仅用于检查过期时间）
function parseJwtPayload(token) {
    try {
        const parts = token.split(".");
        if (parts.length !== 3) return null;
        const payload = JSON.parse(atob(parts[1].replace(/-/g, "+").replace(/_/g, "/")));
        return payload;
    } catch (e) {
        return null;
    }
}

// 检查 Access Token 是否过期（提前 30 秒视为过期）
function isAccessTokenExpired(token) {
    if (!token) return true;
    const payload = parseJwtPayload(token);

    // 如果解析失败或没有 exp 字段，则认为“未过期”，交给后端校验
    if (!payload || !payload.exp) return false;
    const now = Math.floor(Date.now() / 1000);
    return payload.exp <= now + 30;
}

// 从 ResponseData 中提取 access_token（支持 data 为字符串或 { access_token: string }）
function extractAccessToken(data) {
    if (!data || data.code !== CODE_SUCCESS || !data.data) return null;
    if (typeof data.data === "string") return data.data;
    return data.data.access_token || null;
}

// 调用 Refresh 接口刷新 Token
function refreshAccessToken() {
    return fetch("/refreshtoken", {
        method: "POST",
        credentials: "include"
    })
        .then(res => res.json())
        .then(data => {
            var token = extractAccessToken(data);
            if (token) {
                setAccessToken(token);
                return true;
            }
            return false;
        })
        .catch(() => false);
}

// 进入页面时检查认证状态：无 token / 过期则尝试 Refresh
async function checkAuth() {
    const token = getAccessToken();
    const expired = isAccessTokenExpired(token);
    if (!expired && token) return true;
    return await refreshAccessToken();
}

// 根据登录状态更新导航栏
function updateNavbar(isLoggedIn) {
    const authArea = document.getElementById("auth-area");
    const guestArea = document.getElementById("guest-area");
    if (guestArea) guestArea.style.display = isLoggedIn ? "none" : "flex";
    if (authArea) authArea.style.display = isLoggedIn ? "flex" : "none";
}

// 退出登录
function logout() {
    clearAccessToken();
    window.location.href = "/page/HomePage";
}

function addToCart(bookId) {
    const token = getAccessToken();
    const headers = { "Content-Type": "application/json" };
    if (token) headers["Authorization"] = "Bearer " + token;

    fetch("/api/cart", {
        method: "POST",
        headers: headers,
        credentials: "include",
        body: JSON.stringify({ book_id: bookId })
    })
        .then(response => response.json())
        .then(data => {
            if (data.code === CODE_NEED_LOGIN || data.code === CODE_INVALID_TOKEN) {
                if (confirm("请先登录后再操作")) {
                    window.location.href = "/page/LoginPage";
                }
                return;
            }
            if (data.code === CODE_SUCCESS) {
                alert("已加入购物车");
            } else {
                alert(data.msg || "加入购物车失败");
            }
        })
        .catch(error => {
            console.error("错误:", error);
            alert("加入购物车失败");
        });
}