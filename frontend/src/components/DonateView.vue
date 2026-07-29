<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { GetDonateImageURLs } from '../../bindings/sugarplayer/app'

const urls = ref<{ wechat: string; alipay: string }>({ wechat: '', alipay: '' })
const loaded = ref(false)

onMounted(async () => {
  try {
    const res = await GetDonateImageURLs()
    urls.value = {
      wechat: (res && res.wechat) || '',
      alipay: (res && res.alipay) || '',
    }
  } catch {
    urls.value = { wechat: '', alipay: '' }
  }
  loaded.value = true
})
</script>

<template>
  <div class="donate-view">
    <h1 class="title">赞助作者 ♥</h1>
    <p class="subtitle">如果这个播放器对你有帮助，欢迎请作者喝杯咖啡～</p>

    <div class="cards">
      <div v-if="urls.wechat" class="card">
        <div class="qr-wrap">
          <img :src="urls.wechat" alt="微信赞赏" />
        </div>
        <div class="card-name">微信</div>
      </div>
      <div v-if="urls.alipay" class="card">
        <div class="qr-wrap">
          <img :src="urls.alipay" alt="支付宝" />
        </div>
        <div class="card-name">支付宝</div>
      </div>
    </div>

    <p v-if="loaded && !urls.wechat && !urls.alipay" class="empty">
      暂未配置捐赠二维码
    </p>
  </div>
</template>

<style scoped>
.donate-view {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  padding: 32px;
  text-align: center;
  color: var(--fluent-text);
  overflow-y: auto;
}

.title {
  margin: 0;
  font-size: 26px;
  font-weight: 700;
}

.subtitle {
  margin: 0;
  font-size: 14px;
  color: var(--fluent-text-secondary);
}

.cards {
  display: flex;
  gap: 40px;
  margin-top: 12px;
  flex-wrap: wrap;
  justify-content: center;
}

.card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
}

.qr-wrap {
  width: 240px;
  height: 240px;
  padding: 14px;
  border-radius: 16px;
  background: #fff;
  border: 1px solid var(--fluent-border);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
}

.qr-wrap img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
}

.card-name {
  font-size: 15px;
  font-weight: 600;
}

.empty {
  margin-top: 16px;
  font-size: 13px;
  color: var(--fluent-text-secondary);
}
</style>
