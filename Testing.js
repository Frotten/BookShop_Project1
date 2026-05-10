import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {

    scenarios: {

        seckill_test: {

            executor: 'ramping-vus',

            stages: [
                { duration: '10s', target: 300 },
                { duration: '20s', target: 1000 },
                { duration: '30s', target: 2000 },
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

function randomUser() {

    const id = Math.floor(Math.random() * 100000000);

    return {

        username: `user_${id}`,

        password: '123456',

        email: `user_${id}@test.com`,
    };
}

export default function () {

    // ====================================
    // 1. 生成用户
    // ====================================

    const user = randomUser();

    // ====================================
    // 2. 注册
    // ====================================

    const registerPayload = JSON.stringify({

        username: user.username,

        password: user.password,

        re_Password: user.password,

        email: user.email,

        gender: 1,
    });

    const registerRes = http.post(

        `${BASE_URL}/api/register`,

        registerPayload,

        {
            headers: {
                'Content-Type': 'application/json',
            },
        }
    );

    check(registerRes, {

        'register success': (r) => r.status === 200,
    });

    // ====================================
    // 3. 登录
    // ====================================

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

    // ====================================
    // 4. 获取 token
    // ====================================

    const token = loginRes.json('data.access_token');

    check(token, {

        'token exists': (t) => t !== '',
    });

    // ====================================
    // 5. 获取秒杀列表
    // ====================================

    const listRes = http.get(

        `${BASE_URL}/api/seckill/list`
    );

    check(listRes, {

        'get seckill list success': (r) => r.status === 200,
    });

    // ====================================
    // 6. 秒杀
    // ====================================

    const seckillPayload = JSON.stringify({

        product_id: 8,
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

    check(seckillRes, {

        'seckill request success': (r) => r.status === 200,
    });

    // ====================================
    // 7. 查询订单
    // ====================================

    const orderRes = http.get(

        `${BASE_URL}/api/userOrders`,

        {
            headers: {

                Authorization: `Bearer ${token}`,
            },
        }
    );

    check(orderRes, {

        'get orders success': (r) => r.status === 200,
    });

    sleep(1);
}