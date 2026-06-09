<template>
  <div class="chats-container">
    <h1>My Conversations</h1>

    <!-- If there are no conversations at all -->
    <div v-if="conversations.length === 0">
      <p>No conversations yet.</p>
    </div>

    <!-- Otherwise show the conversation list -->
    <ul v-else>
      <li v-for="chat in conversations" :key="chat.id" class="chat-item">
        <RouterLink :to="`/chat/${chat.id}`" class="chat-link">
          <div class="chat-info">
            <img
              v-if="chat.photo"
              :src="chat.photo"
              alt="Chat Photo"
              class="chat-photo"
              @error="setDefaultPhoto($event)"
            />
            <div class="chat-text-info">
              <span class="chat-name">{{ chat.name }}</span>
              <!-- Display the preview for the last message -->
              <p class="last-message">{{ getLastMessagePreview(chat) }}</p>
            </div>
          </div>
          <p class="last-convo-time">{{ formatDate(chat.last_convo) }}</p>
        </RouterLink>
      </li>
    </ul>

    <div class="start-new-convo">
      <RouterLink to="/sendMessageFirstView">
        <button class="new-convo-btn">Start a New Conversation</button>
      </RouterLink>
      <div class="start-new-group">
        <RouterLink to="/createGroupView">
          <button class="new-group-btn">Create a New Group</button>
        </RouterLink>
      </div>
      <RouterLink to="/search/users" class="search-people-button">
        <button class="new-group-btn">Search People</button>
      </RouterLink>
    </div>
  </div>
</template>

<script>
import axios from '../services/axios.js';

export default {
  name: "ConversationsView",
  data() {
    return {
      conversations: [],
      loading: true,
      updateInterval: null, // To store the setInterval reference
    };
  },
  created() {
    // Fetch right away on creation
    this.fetchConversations();
  },
  mounted() {
    // Poll the server every 1 second to mimic real-time updates
    this.updateInterval = setInterval(() => {
      this.fetchConversations();
    }, 1000);
  },
  beforeUnmount() {
    // Clean up the interval to avoid memory leaks
    if (this.updateInterval) {
      clearInterval(this.updateInterval);
    }
  },
  methods: {
    async fetchConversations() {
      const token = localStorage.getItem("authToken");
      const userID = localStorage.getItem("userID");

      if (!token || !userID) {
        console.warn("🚨 User not authenticated.");
        this.loading = false;
        return;
      }

      this.loading = true; // show loading spinner or text if needed
      try {
        console.log("🔍 Fetching user conversations...");
        const response = await axios.get(`/users/${userID}/conversations`, {
          headers: { Authorization: `Bearer ${token}` },
        });

        console.log("📝 Full API Response:", response.data);

        // Process each chat object from the API
        const baseURL = axios.defaults.baseURL;
        const mappedConversations = response.data.map((chat) => {
          let photoURL = "/default-profile.png"; // Default image

          // If a valid photo is provided (not the default), build its URL using baseURL
          if (chat.photo && chat.photo.String && chat.photo.String !== "/default-profile.png") {
            photoURL = chat.photo.String.startsWith(baseURL)
              ? chat.photo.String
              : `${baseURL}${chat.photo.String}`;
          }

          return {
            id: chat.id,
            name: chat.name || "Unnamed Chat",
            photo: photoURL,
            last_convo: chat.last_convo,
            // These fields remain as sql.NullString objects (with .Valid and .String)
            last_message: chat.last_message,
            last_message_type: chat.last_message_type,
          };
        });

        // Sort conversations by last_convo descending (latest first)
        mappedConversations.sort((a, b) => {
          // Convert to Date objects and subtract
          const dateA = new Date(a.last_convo);
          const dateB = new Date(b.last_convo);
          return dateB - dateA; // descending
        });

        // Assign sorted array to local state
        this.conversations = mappedConversations;
        console.log("✅ Processed & Sorted Conversations Data:", this.conversations);
      } catch (error) {
        console.error("❌ Error fetching conversations:", error);
      } finally {
        this.loading = false;
      }
    },
    setDefaultPhoto(event) {
      event.target.src = "/default-profile.png";
    },
    formatDate(dateString) {
      const date = new Date(dateString);
      return date.toLocaleString();
    },
    /**
     * Returns a preview for the last message:
     * - If the type is "text", returns a truncated version (first 20 characters with ellipsis if longer).
     * - If the type is "photo" or "gif", returns the literal string "photo" or "gif".
     * - Otherwise, returns an empty string.
     *
     * This method "unwraps" the sql.NullString objects by checking their .String property.
     */
    getLastMessagePreview(chat) {
      // Unwrap the last message and type from the sql.NullString objects.
      const msg = (chat.last_message && chat.last_message.String) || "";
      const type = (chat.last_message_type && chat.last_message_type.String) || "";

      if (!msg) {
        return "";
      }
      if (type === "text") {
        return msg.length > 20 ? msg.substring(0, 20) + "..." : msg;
      } else if (type === "photo") {
        return "photo";
      } else if (type === "gif") {
        return "gif";
      } else {
        return "";
      }
    },
  },
};
</script>

<style scoped>
/* WhatsApp Color Palette */
.chats-container {
  background-color: #111b21;
  min-height: 100vh;
  padding: 20px;
  color: #e9edef;
}

.chats-container h1 {
  color: #e9edef;
  font-size: 1.5rem;
  font-weight: 500;
  margin-bottom: 20px;
  text-align: center;
}

.chats-container > div > p {
  color: #aebac1;
  text-align: center;
}

.chats-container ul {
  padding: 0;
  margin: 0;
}

.chat-item {
  list-style: none;
  margin: 0;
}

.chat-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border: none;
  border-bottom: 1px solid #2a3942;
  text-decoration: none;
  color: #e9edef;
  background-color: #111b21;
  transition: background-color 0.2s ease;
}

.chat-link:hover {
  background-color: #202c33;
}

.chat-photo {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  object-fit: cover;
  border: none;
}

.chat-info {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.chat-text-info {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  flex: 1;
}

.chat-name {
  font-weight: 500;
  font-size: 1rem;
  color: #e9edef;
}

.last-message {
  font-size: 0.875rem;
  color: #aebac1;
  margin: 0;
}

.last-convo-time {
  font-size: 0.75rem;
  color: #667781;
  padding-left: 10px;
  padding-right: 0;
  align-self: flex-start;
  margin: 0;
  white-space: nowrap;
}

.start-new-convo {
  margin-top: 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  align-items: center;
}

.new-convo-btn {
  padding: 12px 24px;
  background-color: #00a884;
  color: #111b21;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  font-size: 0.95rem;
  transition: all 0.2s ease;
  width: 100%;
  max-width: 300px;
}

.new-convo-btn:hover {
  background-color: #06cf9c;
  transform: translateY(-1px);
}

.new-group-btn {
  padding: 12px 24px;
  background-color: #00a884;
  color: #111b21;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  font-size: 0.95rem;
  transition: all 0.2s ease;
  width: 100%;
  max-width: 300px;
}

.new-group-btn:hover {
  background-color: #06cf9c;
  transform: translateY(-1px);
}

.start-new-group {
  width: 100%;
  display: flex;
  justify-content: center;
}

.search-people-button {
  width: 100%;
  display: flex;
  justify-content: center;
  text-decoration: none;
}

/* Scrollbar */
.chats-container::-webkit-scrollbar {
  width: 6px;
}

.chats-container::-webkit-scrollbar-track {
  background: #111b21;
}

.chats-container::-webkit-scrollbar-thumb {
  background: #374955;
  border-radius: 3px;
}

.chats-container::-webkit-scrollbar-thumb:hover {
  background: #4a5c6a;
}
</style>

