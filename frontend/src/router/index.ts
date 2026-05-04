import { createRouter, createWebHistory } from 'vue-router'
import ProjectDetailView from '@/views/ProjectDetailView.vue'
import ProjectListView from '@/views/ProjectListView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'project-list',
      component: ProjectListView,
    },
    {
      path: '/projects/:name',
      name: 'project-detail',
      component: ProjectDetailView,
      props: true,
    },
  ],
  scrollBehavior() {
    return { top: 0 }
  },
})

export default router
