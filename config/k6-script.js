import http from 'k6/http';

export const options = {
    scenarios: {
        constant_request_rate: {
            executor: 'constant-arrival-rate',
            rate: 50,
            timeUnit: '1s',
            duration: '2h',
            preAllocatedVUs: 10,
            maxVUs: 100,
        },
    },
};

export default function () {
    http.get('http://hello-world:8080');
    // uncomment lines below to generate load on /error and /long handlers
    // http.get('http://hello-world:8080/error');
    // http.get('http://hello-world:8080/long');
}