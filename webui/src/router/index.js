import { createRouter, createWebHashHistory } from 'vue-router';
import HomeView from '../views/HomeView.vue';
import LoginView from "../views/LoginView.vue";
import UsernameView from "../views/UsernameView.vue";
import SetMyPhotoView from "../views/ProfileView.vue";
import ConversationsView from "../views/ConversationsView.vue";
import SendMessageFirstView from "../views/SendMessageFirstView.vue"; // Import the new view
import CreateGroupView from "../views/CreateGroupView.vue"; // Import the new CreateGroupView
import GetMessagesView from '../views/GetMessagesView.vue';
import SearchUserView from "../views/SearchUserView.vue";

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', redirect: '/login' }, // ✅ Redirects to Login correctly
    { path: '/login', component: LoginView }, // ✅ Now login works
    { path: '/home', component: HomeView }, 
    { 
      path: '/users/me/username', 
      component: UsernameView,
      beforeEnter: (to, from, next) => {
        const token = localStorage.getItem("authToken");
        if (!token) next("/login"); else next();
      }
    },
    { 
      path: '/users/me/photo', 
      component: SetMyPhotoView,
      beforeEnter: (to, from, next) => {
        const token = localStorage.getItem("authToken");
        if (!token) next("/login"); else next();
      }
    },
    { 
      path: "/conversations", 
      component: ConversationsView,
      beforeEnter: (to, from, next) => {
        const token = localStorage.getItem("authToken");
        if (!token) next("/login"); else next();
      }
    },
    {
      path: '/sendMessageFirstView', 
      component: SendMessageFirstView,
      beforeEnter: (to, from, next) => {
        const token = localStorage.getItem("authToken");
        if (!token) next("/login"); else next();
      }
    },
	{
		path: '/createGroupView',
		component: CreateGroupView,
		beforeEnter: (to, from, next) => {
		  const token = localStorage.getItem("authToken");
		  if (!token) next("/login"); else next();
		}
	},
	{
		path: '/chat/:c_id',  // :c_id is the dynamic parameter for conversation ID
		name: 'GetMessages',
		component: GetMessagesView,
	},
  {
    path: '/search/users',
    component: SearchUserView,
    beforeEnter: (to, from, next) => {
      const token = localStorage.getItem("authToken");
      if (!token) next("/login");
      else next();
    }
  }

  ]
});

export default router;