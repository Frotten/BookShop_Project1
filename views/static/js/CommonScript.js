function addToCart(bookId) {

    fetch("/api/v1/cart", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            book_id: bookId
        })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error("请求失败");
            }
            return response.json();
        })
        .then(data => {
            alert("已加入购物车");
        })
        .catch(error => {
            console.error("错误:", error);
            alert("加入购物车失败");
        });
}