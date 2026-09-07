import { startAuthentication, startRegistration } from '@simplewebauthn/browser';
import axios from 'axios';
const API_BASE_PATH = process.env.SNTZ_BASE_PATH || 'http://localhost:8080';
async function PasskeyRegisterChallenge(params) {
    try {
        const endpoint = `${API_BASE_PATH}/iam/passkey/register/challenge`;
        const resp = await axios.get(endpoint, {
            params,
        });
        return resp.data;
    }
    catch (error) {
        throw error;
    }
}
async function PasskeyRegisterVerify(params) {
    try {
        const endpoint = `${API_BASE_PATH}/iam/passkey/register/verify`;
        const resp = await axios.post(endpoint, params);
        return resp.data;
    }
    catch (error) {
        throw error;
    }
}
async function PasskeyLoginChallenge(params) {
    try {
        const endpoint = `${API_BASE_PATH}/iam/passkey/login/challenge`;
        const resp = await axios.get(endpoint, { params });
        return resp.data;
    }
    catch (error) {
        throw error;
    }
}
async function PasskeyLoginVerify(params) {
    try {
        const endpoint = `${API_BASE_PATH}/iam/passkey/login/verify`;
        const resp = await axios.put(endpoint, params);
        return resp.data;
    }
    catch (error) {
        throw error;
    }
}
export async function PasskeyRegister(params) {
    try {
        const start = await PasskeyRegisterChallenge(params);
        const options = start.options ? start.options : undefined;
        if (options == undefined) {
            throw new Error('publicKey is null');
        }
        const credential = await startRegistration({
            optionsJSON: options['publicKey'],
        });
        const encoded = btoa(JSON.stringify(credential));
        const finish = await PasskeyRegisterVerify({
            sessionId: start.sessionId,
            credentialCreationResponse: encoded,
        });
        return finish;
    }
    catch (error) {
        throw error;
    }
}
export async function PasskeyLogin(params) {
    try {
        const start = await PasskeyLoginChallenge(params);
        const options = start.options ? start.options : undefined;
        if (options == undefined) {
            throw new Error('publicKey is null');
        }
        const credential = await startAuthentication({
            optionsJSON: options['publicKey'],
        });
        const encoded = btoa(JSON.stringify(credential));
        const finish = await PasskeyLoginVerify({
            sessionId: start.sessionId,
            credentialAssertionData: encoded,
        });
        return finish;
    }
    catch (error) {
        throw error;
    }
}
