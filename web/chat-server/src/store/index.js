import { createStore } from 'vuex'

export default createStore({
  state: {
    // web服务器地址 - Docker环境配置
    backendUrl: 'http://localhost:8000',
    wsUrl: 'ws://localhost:8000',
    // 原始云服务器配置（已注释）
    // backendUrl: 'https://123.56.164.220:8000',
    // wsUrl: 'wss://123.56.164.220:8000',
    // 信令服务器地址
    // signalUrl: 'wss://127.0.0.1:8001',
    userInfo: (sessionStorage.getItem('userInfo') && JSON.parse(sessionStorage.getItem('userInfo'))) || {},
    socket: null,
  },
  getters: {
  },
  mutations: {
    setUserInfo(state, userInfo) {
      state.userInfo = userInfo;
      sessionStorage.setItem('userInfo', JSON.stringify(userInfo));
    },
    cleanUserInfo(state) {
      state.userInfo = {};
      sessionStorage.removeItem('userInfo');
    }
  },
  actions: {
  },
  modules: {
  }
})
