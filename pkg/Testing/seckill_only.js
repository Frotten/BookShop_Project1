import http from 'k6/http';

import { check, sleep } from 'k6';

import { SharedArray } from 'k6/data';

// ========================================
// 读取账号
// 格式：username,password
// ========================================

const users = new SharedArray('users', function () {

    return open('./users.txt')
        .split('\n')
        .filter(line => line.trim() !== '')
        .map(line => {

            const arr = line.split(',');

            return {

                username: arr[0],

                password: arr[1],
            };
        });
});

// ========================================
// 压测配置
// ========================================

export const options = {

    scenarios: {

        seckill_test: {

            executor: 'ramping-vus',

            stages: [
                { duration: '10s', target: 1000 },
                { duration: '20s', target: 3000 },
                { duration: '30s', target: 5000 },
                { duration: '10s', target: 0 },
            ],
        },
    },

    thresholds: {

        http_req_duration: ['p(95)<500'],

        http_req_failed: ['rate<0.01'],
    },
};

const BASE_URL = 'http://127.0.0.1:9090';

// ========================================
// 每个 VU 独立缓存 token
// ========================================

let token = '';

// ========================================
// 登录函数（完全仿照 Testing.js）
// ========================================

function login(user) {

    const loginPayload = JSON.stringify({

        username: user.username,

        password: user.password,
    });

    const loginRes = http.post(

        `${BASE_URL}/api/login`,

        loginPayload,

        {
            headers: {

                'Content-Type': 'application/json',
            },
        }
    );

    check(loginRes, {

        'login success': (r) => r.status === 200,
    });

    const accessToken = loginRes.json('data.access_token');

    check(accessToken, {

        'token exists': (t) => t !== '',
    });

    return accessToken;
}

// ========================================
// 主逻辑
// ========================================

export default function () {

    // ========================================
    // 每个 VU 只登录一次
    // ========================================

    if (!token) {

        const user = users[__VU % users.length];

        token = login(user);

        console.log(`VU ${__VU} login success`);
    }

    // ========================================
    // 秒杀
    // ========================================

    const seckillPayload = JSON.stringify({

        product_id: 13,
    });

    const seckillRes = http.post(

        `${BASE_URL}/api/seckill/do`,

        seckillPayload,

        {
            headers: {

                'Content-Type': 'application/json',

                Authorization: `Bearer ${token}`,
            },
        }
    );

    console.log(seckillRes.body);

    check(seckillRes, {

        'seckill success': (r) => r.status === 200,
    });

    // ========================================
    // 查询订单
    // ========================================

    const orderRes = http.get(

        `${BASE_URL}/api/userOrders`,

        {
            headers: {

                Authorization: `Bearer ${token}`,
            },
        }
    );

    check(orderRes, {

        'order success': (r) => r.status === 200,
    });

    sleep(1);
}