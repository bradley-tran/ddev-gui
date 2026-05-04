import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { installI18n } from './lib/i18n'
import './styles/global.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)
installI18n(app)

app.mount('#app')
