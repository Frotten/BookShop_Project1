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

function apiFetch(url, options = {}) {
    const token = localStorage.getItem("access_token");

    return fetch(url, {
        ...options,
        headers: {
            "Content-Type": "application/json",
            "Authorization": "Bearer " + token,
            ...(options.headers || {})
        },
        credentials: "include"
    });
}

function addToCart(bookId) {
    const token = getAccessToken();
    const headers = { "Content-Type": "application/json" };
    if (token) headers["Authorization"] = "Bearer " + token;
    apiFetch("/api/cart", {
        method: "POST",
        credentials: "include",
        body: JSON.stringify({book_id: bookId})
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

// 显示数量选择弹窗
function showQuantityDialog(bookId, bookTitle, bookPrice, maxStock) {
    // 检查登录状态
    const token = getAccessToken();
    if (!token || isAccessTokenExpired(token)) {
        if (confirm('请先登录后再操作')) {
            window.location.href = '/page/LoginPage';
        }
        return;
    }
    
    // 创建弹窗遮罩层
    const overlay = document.createElement('div');
    overlay.className = 'dialog-overlay';
    overlay.id = 'quantity-dialog-overlay';
    
    // 创建弹窗内容
    const dialog = document.createElement('div');
    dialog.className = 'quantity-dialog';
    
    const priceYuan = (bookPrice / 100).toFixed(2);
    const stock = maxStock || 99;
    
    dialog.innerHTML = `
        <div class="dialog-header">
            <h3><i class="ri-shopping-cart-2-line"></i> 加入购物车</h3>
            <button class="dialog-close" onclick="closeQuantityDialog()">
                <i class="ri-close-line"></i>
            </button>
        </div>
        <div class="dialog-body">
            ${bookTitle ? `<div class="book-info"><span class="book-title">${escapeHtmlForDialog(bookTitle)}</span></div>` : ''}
            ${bookPrice ? `<div class="book-price">单价：￥${priceYuan}</div>` : ''}
            <div class="quantity-selector">
                <label>购买数量：</label>
                <div class="quantity-control">
                    <button type="button" class="qty-btn qty-minus" onclick="changeQuantity(-1)">
                        <i class="ri-subtract-line"></i>
                    </button>
                    <input type="number" id="dialog-quantity-input" value="1" min="1" max="${stock}" oninput="handleManualInput(event)">
                    <button type="button" class="qty-btn qty-plus" onclick="changeQuantity(1)">
                        <i class="ri-add-line"></i>
                    </button>
                </div>
                <div class="quantity-tips">库存：${stock} 件</div>
            </div>
            <div class="total-price">
                <span>总计：</span>
                <span id="dialog-total-price" class="price-amount">￥${priceYuan}</span>
            </div>
        </div>
        <div class="dialog-footer">
            <button class="btn-cancel" onclick="closeQuantityDialog()">
                <i class="ri-close-circle-line"></i> 取消
            </button>
            <button class="btn-confirm" onclick="confirmAddToCart(${bookId})">
                <i class="ri-check-circle-line"></i> 确认加入购物车
            </button>
        </div>
    `;
    
    overlay.appendChild(dialog);
    document.body.appendChild(overlay);
    
    // 阻止背景滚动
    document.body.style.overflow = 'hidden';
    
    // 绑定 ESC 键关闭
    document.addEventListener('keydown', handleEscKey);
}

// 全局变量存储当前选择的数量
let currentSelectedQuantity = 1;

// 增减数量
function changeQuantity(delta) {
    const input = document.getElementById('dialog-quantity-input');
    if (!input) return;
    
    let newValue = parseInt(input.value) + delta;
    const min = parseInt(input.min) || 1;
    const max = parseInt(input.max) || 99;
    
    if (newValue < min) newValue = min;
    if (newValue > max) newValue = max;
    
    input.value = newValue;
    currentSelectedQuantity = newValue;
    
    // 更新总价
    updateTotalPrice();
}

// 更新总价显示
function updateTotalPrice() {
    const priceElement = document.querySelector('.book-price');
    const totalElement = document.getElementById('dialog-total-price');
    
    if (!priceElement || !totalElement) return;
    
    const priceText = priceElement.textContent.replace('单价：￥', '');
    const price = parseFloat(priceText);
    
    if (!isNaN(price)) {
        const total = (price * currentSelectedQuantity).toFixed(2);
        totalElement.textContent = `￥${total}`;
    }
}

// 关闭弹窗
function closeQuantityDialog() {
    const overlay = document.getElementById('quantity-dialog-overlay');
    if (overlay) {
        overlay.remove();
        document.body.style.overflow = '';
        document.removeEventListener('keydown', handleEscKey);
    }
}

// ESC 键关闭弹窗
function handleEscKey(e) {
    if (e.key === 'Escape') {
        closeQuantityDialog();
    }
}

// 确认加入购物车
function confirmAddToCart(bookId) {
    const quantity = currentSelectedQuantity;
    
    if (quantity < 1) {
        alert('数量至少为 1');
        return;
    }
    
    const token = getAccessToken();
    const headers = { "Content-Type": "application/json" };
    if (token) headers["Authorization"] = "Bearer " + token;
    
    apiFetch("/api/cart", {
        method: "POST",
        credentials: "include",
        body: JSON.stringify({ 
            book_id: bookId,
            quantity: quantity
        })
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
            closeQuantityDialog();
            showSuccessToast(`已成功添加 ${quantity} 件商品到购物车`);
        } else {
            alert(data.msg || "加入购物车失败");
        }
    })
    .catch(error => {
        console.error("错误:", error);
        alert("加入购物车失败");
    });
}

// HTML 转义函数（用于弹窗）
function escapeHtmlForDialog(text) {
    if (!text) return '';
    return String(text)
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}

// 显示成功提示 Toast
function showSuccessToast(message) {
    const toast = document.createElement('div');
    toast.className = 'success-toast';
    toast.innerHTML = `
        <i class="ri-checkbox-circle-fill"></i>
        <span>${escapeHtmlForDialog(message)}</span>
    `;
    
    document.body.appendChild(toast);
    
    setTimeout(() => {
        toast.classList.add('toast-hide');
        setTimeout(() => toast.remove(), 300);
    }, 2000);
}

// ========== 购物车相关功能 ==========

// 获取购物车列表
function getCart() {
    return apiFetch("/api/cart", { method: "GET" })
        .then(response => response.json())
        .then(data => {
            if (data.code === CODE_SUCCESS) {
                return data.data || [];
            } else if (data.code === CODE_NEED_LOGIN || data.code === CODE_INVALID_TOKEN) {
                throw new Error("请先登录");
            } else {
                throw new Error(data.msg || "获取购物车失败");
            }
        })
        .catch(error => {
            console.error("获取购物车错误:", error);
            throw error;
        });
}

// 更新购物车商品数量（带库存验证）
function updateCartItem(bookId, quantity, maxStock) {
    if (quantity <= 0) {
        return removeFromCart(bookId);
    }
    
    if (maxStock && quantity > maxStock) {
        alert(`库存不足，最多可购买 ${maxStock} 件`);
        return Promise.resolve(false);
    }
    
    return apiFetch("/api/cart", {
        method: "PUT",
        credentials: "include",
        body: JSON.stringify({ 
            book_id: bookId,
            quantity: quantity 
        })
    })
    .then(response => response.json())
    .then(data => {
        if (data.code === CODE_SUCCESS) {
            return true;
        } else if (data.code === CODE_NEED_LOGIN || data.code === CODE_INVALID_TOKEN) {
            if (confirm("请先登录后再操作")) {
                window.location.href = "/page/LoginPage";
            }
            return false;
        } else {
            alert(data.msg || "更新失败");
            return false;
        }
    })
    .catch(error => {
        console.error("更新购物车错误:", error);
        alert("更新失败，请重试");
        return false;
    });
}

// 从购物车移除商品
function removeFromCart(bookId) {
    return apiFetch("/api/cart/" + bookId, {
        method: "DELETE",
        credentials: "include"
    })
    .then(response => response.json())
    .then(data => {
        if (data.code === CODE_SUCCESS) {
            showSuccessToast("已从购物车移除");
            return true;
        } else if (data.code === CODE_NEED_LOGIN || data.code === CODE_INVALID_TOKEN) {
            if (confirm("请先登录后再操作")) {
                window.location.href = "/page/LoginPage";
            }
            return false;
        } else {
            alert(data.msg || "删除失败");
            return false;
        }
    })
    .catch(error => {
        console.error("删除购物车错误:", error);
        alert("删除失败，请重试");
        return false;
    });
}

// 清空购物车
function clearCart() {
    return apiFetch("/api/cart", {
        method: "DELETE",
        credentials: "include"
    })
    .then(response => response.json())
    .then(data => {
        if (data.code === CODE_SUCCESS) {
            showSuccessToast("购物车已清空");
            return true;
        } else if (data.code === CODE_NEED_LOGIN || data.code === CODE_INVALID_TOKEN) {
            if (confirm("请先登录后再操作")) {
                window.location.href = "/page/LoginPage";
            }
            return false;
        } else {
            alert(data.msg || "清空失败");
            return false;
        }
    })
    .catch(error => {
        console.error("清空购物车错误:", error);
        alert("清空失败，请重试");
        return false;
    });
}

// 渲染购物车简化列表（用于个人中心等页面）
function renderCartSimpleList(items, containerId) {
    const container = document.getElementById(containerId);
    if (!container) {
        console.error('容器元素不存在:', containerId);
        return;
    }
    
    if (!items || items.length === 0) {
        container.innerHTML = `
            <div class="empty-placeholder">
                <i class="ri-shopping-cart-line"></i>
                <p>购物车空空如也，快去添加喜欢的书籍吧~</p>
                <a href="/page/HomePage">
                    <button style="margin-top:1rem;">
                        <i class="ri-add-line"></i> 去购物
                    </button>
                </a>
            </div>
        `;
        return;
    }
    
    let totalPrice = 0;
    let html = `<div style="margin-bottom:1rem;"><strong>🛒 购物车商品 (${items.length})</strong></div>`;
    
    items.forEach(item => {
        const priceYuan = (item.price / 100).toFixed(2);
        const subtotal = (item.price * item.quantity / 100).toFixed(2);
        totalPrice += parseFloat(subtotal);
        
        html += `
            <div class="cart-item-simple" data-book-id="${item.book_id}" data-stock="${item.stock || 99}">
                <div class="item-header">
                    <span class="book-title" data-book-id="${item.book_id}" onclick="gotoBookDetail(${item.book_id})">
                        📚 ${escapeHtmlForDialog(item.title)}
                    </span>
                    <span class="meta-text">单价: ￥${priceYuan} × ${item.quantity}</span>
                </div>
                <div class="content-text">
                    小计: <span class="price-amount">￥${subtotal}</span>
                    ${item.stock ? `<span style="margin-left:1rem;font-size:0.75rem;">库存: ${item.stock}</span>` : ''}
                </div>
                <div class="cart-actions">
                    <button class="small-link" onclick="handleCartQuantityChange(${item.book_id}, -1)">
                        <i class="ri-subtract-line"></i> 减一
                    </button>
                    <button class="small-link" onclick="handleCartQuantityChange(${item.book_id}, 1)">
                        <i class="ri-add-line"></i> 加一
                    </button>
                    <button class="small-link" onclick="handleRemoveCartItem(${item.book_id})">
                        <i class="ri-delete-bin-line"></i> 删除
                    </button>
                </div>
            </div>
        `;
    });
    
    html += `
        <div style="margin-top:1.5rem; text-align:right; border-top:1px solid var(--gray-200); padding-top:1rem;">
            <strong>总计：<span class="price-amount" style="font-size:1.3rem;">￥${totalPrice.toFixed(2)}</span></strong>
            <div style="margin-top:0.8rem;">
                <button onclick="window.location.href='/page/CartPage'" style="background: var(--primary);">
                    前往购物车详细结算
                </button>
            </div>
        </div>
    `;
    
    container.innerHTML = html;
}

// 处理购物车数量变化（增减）
async function handleCartQuantityChange(bookId, delta) {
    const itemElement = document.querySelector(`.cart-item-simple[data-book-id="${bookId}"]`);
    if (!itemElement) {
        console.error('未找到商品元素');
        return;
    }
    
    const stock = parseInt(itemElement.dataset.stock) || 99;
    const metaText = itemElement.querySelector('.meta-text').textContent;
    const currentQtyMatch = metaText.match(/×\s*(\d+)/);
    const currentQty = currentQtyMatch ? parseInt(currentQtyMatch[1]) : 1;
    
    const newQuantity = currentQty + delta;
    
    if (newQuantity <= 0) {
        if (confirm('确定要移除这件商品吗？')) {
            await handleRemoveCartItem(bookId);
        }
        return;
    }
    
    if (newQuantity > stock) {
        alert(`库存不足，最多可购买 ${stock} 件`);
        return;
    }
    
    const success = await updateCartItem(bookId, newQuantity, stock);
    if (success) {
        showSuccessToast('数量已更新');
        // 触发刷新事件，让调用方自行决定如何刷新
        document.dispatchEvent(new CustomEvent('cartUpdated', { detail: { bookId, newQuantity } }));
    }
}

// 处理删除购物车商品
async function handleRemoveCartItem(bookId) {
    const success = await removeFromCart(bookId);
    if (success) {
        document.dispatchEvent(new CustomEvent('cartUpdated', { detail: { bookId, removed: true } }));
    }
}

// 加载并渲染购物车简化列表（通用方法）
async function loadAndRenderCartSimple(containerId) {
    try {
        const cartItems = await getCart();
        renderCartSimpleList(cartItems, containerId);
        return cartItems;
    } catch (error) {
        console.error('加载购物车失败:', error);
        const container = document.getElementById(containerId);
        if (container) {
            if (error.message === '请先登录') {
                container.innerHTML = `
                    <div class="empty-placeholder">
                        <i class="ri-lock-line"></i>
                        <p>请先登录后查看购物车</p>
                        <a href="/page/LoginPage">
                            <button>去登录</button>
                        </a>
                    </div>
                `;
            } else {
                container.innerHTML = `
                    <div class="empty-placeholder">
                        <i class="ri-error-warning-line"></i>
                        <p>${escapeHtmlForDialog(error.message)}</p>
                        <button onclick="loadAndRenderCartSimple('${containerId}')">重试</button>
                    </div>
                `;
            }
        }
        throw error;
    }
}
