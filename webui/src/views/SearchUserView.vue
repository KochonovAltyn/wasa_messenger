<template>
  <div class="search-user-container">
    <h1 class="title">Search for a User</h1>
    <div class="search-bar">
      <input
        type="text"
        v-model="searchQuery"
        placeholder="Enter username"
        class="search-input"
      />
      <button @click="searchUser" class="search-button">Search</button>
    </div>

    <div v-if="errorMessage" class="error-message">
      {{ errorMessage }}
    </div>

    <div v-if="foundUser" class="user-info">
      <p class="user-name"><strong>{{ foundUser.username }}</strong></p>
      <img
        v-if="foundUser.photo && foundUser.photo.String"
        :src="fullPhotoUrl(foundUser.photo.String)"
        alt="User Photo"
        class="user-photo"
      />
      <div class="conversation-info" v-if="conversationId">
        <p>You already have a conversation with this user.</p>
        <button @click="goToConversation" class="conversation-button">
          Go to Conversation
        </button>
      </div>
      <div class="conversation-info" v-else>
        <p>No conversation exists. Click to start a new conversation.</p>
        <button @click="startConversation" class="conversation-button">
          Start Conversation
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import axios from "../services/axios.js";

export default {
  name: "SearchUserView",
  data() {
    return {
      searchQuery: "",
      foundUser: null,
      conversationId: null,
      errorMessage: ""
    };
  },
  methods: {
    fullPhotoUrl(photoPath) {
      return axios.defaults.baseURL + photoPath;
    },
    async searchUser() {
      this.errorMessage = "";
      this.foundUser = null;
      this.conversationId = null;

      const token = localStorage.getItem("authToken");
      if (!token) {
        this.errorMessage = "User not authenticated.";
        return;
      }
      try {
        const response = await axios.get(
          `/search/users?username=${encodeURIComponent(this.searchQuery)}`,
          {
            headers: { Authorization: `Bearer ${token}` }
          }
        );
        this.foundUser = response.data.user;
        this.conversationId = response.data.conversation_id;
      } catch (err) {
        if (err.response && err.response.status === 404) {
          this.errorMessage = "No such user found.";
        } else {
          this.errorMessage = "Error searching for user.";
        }
      }
    },
    goToConversation() {
      this.$router.push(`/chat/${this.conversationId}`);
    },
    async startConversation() {
      const token = localStorage.getItem("authToken");
      const userID = localStorage.getItem("userID");
      if (!token || !userID) {
        this.errorMessage = "User not authenticated.";
        return;
      }
      try {
        const formData = new FormData();
        formData.append("recipient_username", this.foundUser.username);
        formData.append("content_type", "text");
        formData.append("content", "Hi!");
        const response = await axios.post(
          `/users/${userID}/conversations/first-message`,
          formData,
          {
            headers: {
              Authorization: `Bearer ${token}`,
              "Content-Type": "multipart/form-data"
            }
          }
        );
        if (response.data && response.data.c_id) {
          this.$router.push(`/chat/${response.data.c_id}`);
        }
      } catch {
        this.errorMessage = "Error starting conversation.";
      }
    }
  }
};
</script>

<style scoped>
/* WhatsApp Color Palette */
.search-user-container {
  max-width: 600px;
  margin: 30px auto;
  padding: 32px 24px;
  background: #202c33;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
  text-align: center;
}

.title {
  font-size: 2em;
  margin-bottom: 24px;
  color: #e9edef;
  font-weight: 500;
}

.search-bar {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-bottom: 24px;
  gap: 12px;
}

.search-input {
  flex: 1;
  max-width: 400px;
  padding: 12px 18px;
  font-size: 1em;
  border: 1px solid #2a3942;
  border-radius: 24px;
  outline: none;
  transition: all 0.3s ease;
  background-color: #2a3942;
  color: #e9edef;
}

.search-input::placeholder {
  color: #667781;
}

.search-input:focus {
  border-color: #00a884;
  background-color: #1f2c33;
}

.search-button {
  padding: 12px 24px;
  font-size: 1em;
  background-color: #00a884;
  color: #111b21;
  border: none;
  border-radius: 24px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s ease;
}

.search-button:hover {
  background-color: #06cf9c;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 168, 132, 0.4);
}

.search-button:active {
  transform: translateY(0);
}

.error-message {
  color: #ea4335;
  background-color: rgba(234, 67, 53, 0.15);
  border: 1px solid #ea4335;
  border-radius: 8px;
  padding: 12px 20px;
  font-weight: 500;
  margin-bottom: 20px;
}

.user-info {
  border-top: 1px solid #2a3942;
  padding-top: 24px;
  margin-top: 24px;
}

.user-name {
  font-size: 1.5em;
  color: #e9edef;
  margin-bottom: 20px;
  font-weight: 500;
}

.user-photo {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  object-fit: cover;
  margin-bottom: 20px;
  border: 4px solid #00a884;
  box-shadow: 0 4px 12px rgba(0, 168, 132, 0.3);
}

.conversation-info {
  margin-top: 24px;
}

.conversation-info p {
  font-size: 1em;
  color: #aebac1;
  margin-bottom: 16px;
}

.conversation-button {
  padding: 12px 28px;
  font-size: 1em;
  background-color: #00a884;
  color: #111b21;
  border: none;
  border-radius: 24px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s ease;
}

.conversation-button:hover {
  background-color: #06cf9c;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 168, 132, 0.4);
}

.conversation-button:active {
  transform: translateY(0);
}
</style>