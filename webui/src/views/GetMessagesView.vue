<template>
  <div class="page-wrapper">
    <!-- ✅ ШАПКА ЧАТА -->
    <div class="chat-header">
      <div class="header-left">
        <button class="back-button" @click="$router.push('/conversations')">←</button>

        <img
          :src="conversationPhoto"
          alt="Chat Photo"
          class="header-photo"
          @error="setDefaultHeaderPhoto"
        />

        <div class="header-info">
          <h2 class="header-title">{{ conversationName || "Loading..." }}</h2>
          <p class="header-subtitle" v-if="isGroup">{{ onlineCount }} online</p>
        </div>
      </div>

      <div class="header-actions">
        <!-- ❌ Убрали ⋮ из шапки -->
      </div>
    </div>

    <!-- ✅ Add User Modal -->
    <transition name="modal">
      <div v-if="showAddUserForm" class="modal-overlay" @click="cancelAddUser">
        <div class="modal-form" @click.stop>
          <h3>Add User to Group</h3>
          <input
            type="text"
            placeholder="Enter username"
            v-model="usernameToAdd"
            @keyup.enter="addUserToGroup"
          />
          <div class="modal-buttons">
            <button @click="addUserToGroup" class="primary-btn">Add</button>
            <button @click="cancelAddUser" class="secondary-btn">Cancel</button>
          </div>
        </div>
      </div>
    </transition>

    <!-- ✅ Set Name Modal -->
    <transition name="modal">
      <div v-if="showSetNameForm" class="modal-overlay" @click="cancelSetName">
        <div class="modal-form" @click.stop>
          <h3>Change Group Name</h3>
          <input
            type="text"
            placeholder="Enter new group name"
            v-model="newGroupName"
            @keyup.enter="updateGroupName"
          />
          <div class="modal-buttons">
            <button @click="updateGroupName" class="primary-btn">Save</button>
            <button @click="cancelSetName" class="secondary-btn">Cancel</button>
          </div>
        </div>
      </div>
    </transition>

    <!-- ✅ Set Photo Modal -->
    <transition name="modal">
      <div v-if="showSetPhotoForm" class="modal-overlay" @click="cancelSetPhoto">
        <div class="modal-form" @click.stop>
          <h3>Change Group Photo</h3>
          <input
            type="file"
            accept="image/*"
            @change="handleGroupPhotoChange"
            class="file-input-modal"
          />
          <div class="modal-buttons">
            <button
              @click="updateGroupPhoto"
              class="primary-btn"
              :disabled="!newGroupPhotoFile"
            >
              Save
            </button>
            <button @click="cancelSetPhoto" class="secondary-btn">Cancel</button>
          </div>
        </div>
      </div>
    </transition>

    <!-- ✅ Messages Container -->
    <div class="messages-container" ref="messagesContainer">
      <div v-if="loading" class="loading-message">Loading messages...</div>

      <div v-else-if="messages.length === 0" class="empty-message">
        <p>There are no messages yet</p>
      </div>

      <ul v-else>
        <li v-for="message in messages" :key="message.id" class="message-item">
          <div class="message-info">
            <img
              v-if="message.sender_photo && message.sender_photo.String"
              :src="getImageUrl(message.sender_photo.String)"
              alt="Sender Photo"
              class="sender-photo"
            />

            <div class="message-content">
              <div class="message-header">
                <span class="sender-name">{{ message.sender_username }}</span>
                <span v-if="message.status === 'forwarded'" class="forwarded-label"
                  >Forwarded</span
                >
                <span class="message-time">{{ formatDate(message.datetime) }}</span>

                <span
                  v-if="message.sender_id === currentUser"
                  class="message-status"
                >
                  <span
                    v-if="!message.status || message.status === 'sent'"
                    class="status-sent"
                    >✓</span
                  >
                  <span
                    v-else-if="message.status === 'delivered'"
                    class="status-delivered"
                    >✓✓</span
                  >
                  <span
                    v-else-if="message.status === 'read'"
                    class="status-read"
                    >✓✓</span
                  >
                </span>
              </div>

              <div class="message-actions">
                <button class="reply-button" @click="initiateReply(message)">
                  Reply
                </button>
                <button class="forward-button" @click="toggleForwardPanel(message)">
                  {{ message.showForwardPanel ? "Cancel" : "Forward" }}
                </button>
                <button class="comment-button" @click="toggleComments(message)">
                  {{ message.showComments ? "Hide" : "Comment" }}
                </button>
                <button
                  v-if="message.sender_id === currentUser"
                  @click="deleteMessage(message.id)"
                  class="delete-button"
                >
                  Delete
                </button>
              </div>

              <div v-if="hasReply(message)" class="reply-snippet">
                <strong>Replying to: {{ getReplySender(message.reply_to_sender) }}</strong>
                <br />
                <em>"{{ replySnippet(message.reply_to_content) }}"</em>
              </div>

              <div v-if="isImage(message.content)">
                <img
                  :src="getImageUrl(message.content)"
                  alt="Image Message"
                  class="message-media"
                />
              </div>
              <div v-else-if="isGif(message.content)">
                <img
                  :src="getImageUrl(message.content)"
                  alt="Gif Message"
                  class="message-media"
                />
              </div>
              <div v-else>
                <p class="message-text">{{ message.content }}</p>
              </div>

              <div v-if="message.showForwardPanel" class="forward-panel">
                <input
                  type="text"
                  v-model="message.forwardTarget"
                  placeholder="Username or group"
                  @keyup.enter="forwardMessageHandler(message)"
                />
                <button @click="forwardMessageHandler(message)">Forward</button>
              </div>

              <div v-if="message.showComments" class="comments-section">
                <div v-for="c in message.comments" :key="c.id" class="single-comment">
                  <p class="comment-header">
                    <strong>{{ c.username }}</strong>
                    <span class="comment-time">{{ formatDate(c.timestamp) }}</span>
                    <button
                      v-if="c.user_id === currentUser"
                      @click="deleteComment(message, c)"
                      class="delete-comment-button"
                    >
                      Delete
                    </button>
                  </p>
                  <p class="comment-text">{{ c.content }}</p>
                </div>
                <div class="add-comment-form">
                  <div class="reactions-bar">
  <button
    v-for="e in reactionEmojis"
    :key="e"
    class="reaction-btn"
    @click="addEmojiComment(message, e)"
  >
    {{ e }}
  </button>
</div>

                </div>
              </div>
            </div>
          </div>
        </li>
      </ul>
    </div>

    <!-- ✅ Replying Bar -->
    <transition name="slide-up">
      <div v-if="replyToMessage" class="replying-bar">
        <div>
          <strong>Replying to: {{ replyToMessage.sender_username }}</strong>
          <br />
          <em>{{ replySnippet(replyToMessage.content) }}</em>
        </div>
        <button class="cancel-reply" @click="cancelReply">✕</button>
      </div>
    </transition>

    <!-- ✅ Панель действий (снизу, без dropdown) -->
    <transition name="slide-up">
      <div v-if="isGroup && showBottomActions" class="bottom-actions" @click.stop>
        <button @click="handleAddUser">👤 Add User</button>
        <button @click="handleChangeName">✏️ Change Name</button>
        <button @click="handleChangePhoto">📷 Change Photo</button>
        <button class="danger" @click="handleLeaveGroup">🚪 Leave Group</button>
      </div>
    </transition>

    <!-- ✅ Message Input Bar -->
    <div class="message-input-bar">
      <textarea
        v-model="messageText"
        placeholder="Type a message..."
        class="message-textarea"
        @keydown.enter.exact.prevent="sendMessage"
        rows="1"
      ></textarea>

      <label class="file-upload-label">
        📎
        <input
          type="file"
          accept="image/*, .gif"
          class="file-input"
          @change="handleFileChange"
        />
      </label>

      <!-- ✅ ⋮ ВНИЗУ -->
      <button
        v-if="isGroup"
        class="more-btn"
        title="More"
        @click.stop="toggleBottomActions"
      >
        ⋮
      </button>

      <button
        @click="sendMessage"
        class="send-button"
        :disabled="!messageText.trim() && !selectedFile"
      >
        Send
      </button>
    </div>

    <!-- ✅ кликаешь в любое пустое место — закрываются нижние действия -->
    <div v-if="showBottomActions" class="bottom-actions-backdrop" @click="showBottomActions=false"></div>
  </div>
</template>

<script>
import axios from "../services/axios.js";

export default {
  name: "GetMessagesView",
  data() {
    return {
      currentUser: localStorage.getItem("userID"),
      messages: [],
      loading: true,
      isGroup: false,

      messageText: "",
      selectedFile: null,

      isInteracting: false,
      reloadInterval: null,

      showAddUserForm: false,
      usernameToAdd: "",

      showSetNameForm: false,
      newGroupName: "",

      showSetPhotoForm: false,
      newGroupPhotoFile: null,

      replyToMessage: null,

      conversationName: "Loading...",
      conversationPhoto: "/default-profile.png",
      onlineCount: 0,

      // ✅ новое
      showBottomActions: false,
      reactionEmojis: ["❤️", "😂", "😮", "😢", "😡", "👍"],

    };
  },

  async created() {
    await this.getConversation();
  },

  mounted() {
    this.$nextTick(() => this.scrollToBottom());

    this.reloadInterval = setInterval(() => {
      // не обновляем, пока пользователь что-то делает или открыта панель действий
      if (!this.isInteracting && !this.showBottomActions) {
        this.getConversation();
      }
    }, 1000);
  },

  beforeUnmount() {
    if (this.reloadInterval) clearInterval(this.reloadInterval);
  },

  methods: {
    // ====== UI helpers ======
    toggleBottomActions() {
      this.showBottomActions = !this.showBottomActions;
      this.isInteracting = this.showBottomActions;
    },

    setDefaultHeaderPhoto(event) {
      event.target.src = "/default-profile.png";
    },

    scrollToBottom() {
      const container = this.$refs.messagesContainer;
      if (container) container.scrollTop = container.scrollHeight;
    },

    formatDate(dateString) {
      const date = new Date(dateString);
      const now = new Date();
      const diff = now - date;
      const hours = Math.floor(diff / (1000 * 60 * 60));

      if (hours < 24) {
        return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
      }
      return date.toLocaleDateString();
    },

    isImage(content) {
      return /\.(jpg|jpeg|png)$/i.test(content);
    },

    isGif(content) {
      return /\.(gif)$/i.test(content);
    },

    handleFileChange(event) {
      this.selectedFile = event.target.files[0];
      this.isInteracting = true;
    },

    initiateReply(message) {
      this.replyToMessage = message;
      this.isInteracting = true;
    },

    cancelReply() {
      this.replyToMessage = null;
      this.isInteracting = false;
    },

    snippetOf(text) {
      if (!text) return "";
      return text.length > 40 ? text.substring(0, 40) + "…" : text;
    },

    getPlainValue(val) {
      if (val && typeof val === "object" && "String" in val) return val.String;
      return val;
    },

    hasReply(message) {
      const sender = this.getPlainValue(message.reply_to_sender);
      const content = this.getPlainValue(message.reply_to_content);
      return message.reply_to && (sender || content);
    },

    getReplySender(val) {
      return this.getPlainValue(val);
    },

    replySnippet(content) {
      const plain = this.getPlainValue(content);
      if (!plain) return "";
      if (this.isGif(plain)) return "gif";
      if (this.isImage(plain)) return "photo";
      return this.snippetOf(plain);
    },

    // ====== Bottom actions handlers (используем твои модалки/функции) ======
    handleAddUser() {
      this.showBottomActions = false;
      this.showAddUserForm = true;
      this.isInteracting = true;
    },

    handleChangeName() {
      this.showBottomActions = false;
      this.showSetNameForm = true;
      this.isInteracting = true;
    },

    handleChangePhoto() {
      this.showBottomActions = false;
      this.showSetPhotoForm = true;
      this.isInteracting = true;
    },

    handleLeaveGroup() {
      this.showBottomActions = false;
      if (confirm("Are you sure you want to leave this group?")) {
        this.leaveGroup();
      }
    },

    // ====== API ======
    async getConversation() {
      const token = localStorage.getItem("authToken");
      const conversationID = this.$route.params.c_id; // id из чата url 

      if (!token || !conversationID) {
        console.warn("Missing token or conversation ID.");
        return;
      }

      try {
        const response = await axios.get(`conversations/${conversationID}`, {
          headers: { Authorization: `Bearer ${token}` },
        });

        const fetchedMessages = Array.isArray(response.data.messages)
          ? response.data.messages
          : [];

        const existingMessagesMap = new Map(this.messages.map((m) => [m.id, m]));
        // ВАЖНО: Мержим с существующими, чтобы не потерять локальное состояние
        const mergedMessages = fetchedMessages.map((fetchedMsg) => {
          const existingMsg = existingMessagesMap.get(fetchedMsg.id);

          if (existingMsg) {
            return {
              ...fetchedMsg,
              showComments: existingMsg.showComments,
              showForwardPanel: existingMsg.showForwardPanel,
              forwardTarget: existingMsg.forwardTarget,
              status: existingMsg.status,
              reply_to: existingMsg.reply_to,
              reply_to_content: existingMsg.reply_to_content,
              reply_to_sender: existingMsg.reply_to_sender,
              comments: existingMsg.comments,
            };
          }
          // новое сообщение - инициализируем
          return {
            ...fetchedMsg,
            comments: [],
            showComments: false,
            showForwardPanel: false,
            forwardTarget: "",
            status: fetchedMsg.status || "",
          };
        });

        this.messages = mergedMessages;

        const conv = response.data.conversation;
        this.isGroup = !!conv.is_group;

        // Для групп используем имя группы, для личных чатов - имя собеседника
        if (this.isGroup) {
          this.conversationName = conv.name || "Group Chat";
        } else {
          // Для личного чата берем имя из other_user_username
          // Проверяем все возможные варианты структуры данных
          const otherUsername = conv.other_user_username?.String 
            || conv.other_user_username 
            || conv.other_username?.String 
            || conv.other_username
            || conv.name?.String
            || conv.name;
          
          this.conversationName = otherUsername || "Chat";
          console.log("Conv data:", conv);
          console.log("Setting conversation name to:", this.conversationName);
        }
        // аватар чата 
        const baseURL = axios.defaults.baseURL;
        if (conv.photo && conv.photo.String && conv.photo.String !== "/default-profile.png") {
          this.conversationPhoto = conv.photo.String.startsWith(baseURL)
            ? conv.photo.String
            : `${baseURL}${conv.photo.String}`;
        } else {
          this.conversationPhoto = "/default-profile.png";
        }

        this.onlineCount = Math.floor(Math.random() * 10) + 1;
      } catch (error) {
        console.error("Error fetching conversation:", error);
      } finally {
        this.loading = false;
      }
    },

    async sendMessage() {
      if (!this.messageText.trim() && !this.selectedFile) return; // нельзя отправить пустое 

      const token = localStorage.getItem("authToken");
      const conversationID = this.$route.params.c_id;
      if (!token || !conversationID) return;

      try {
        const formData = new FormData();
        formData.append("content", this.messageText);
        formData.append("content_type", "text");
        // если это ответ на сообщение 
        if (this.replyToMessage) formData.append("reply_to", this.replyToMessage.id.toString());

        if (this.selectedFile) {
          formData.append("file", this.selectedFile);
          formData.append("content_type", "photo");
        }
        // выше если прикпрлен файл 
        const response = await axios.post(
          `/conversations/${conversationID}/messages`,
          formData,
          {
            headers: {
              Authorization: `Bearer ${token}`,
              "Content-Type": "multipart/form-data",
            },
          }
        );
        // Добавляем сообщение локально (для мгновенного отображения)
        const baseURL = axios.defaults.baseURL;
        const newMessage = {
          id: response.data.message_id || Date.now(),
          content: response.data.content,
          sender_username: response.data.sender_username,
          sender_photo: response.data.sender_photo
            ? (response.data.sender_photo.startsWith(baseURL)
                ? response.data.sender_photo
                : baseURL + response.data.sender_photo)
            : "/default-profile.png",
          datetime: new Date().toISOString(),
          sender_id: this.currentUser,
          status: "",
          comments: [],
          showComments: false,
          showForwardPanel: false,
          forwardTarget: "",
          reply_to: this.replyToMessage ? this.replyToMessage.id : null,
          reply_to_sender: this.replyToMessage ? this.replyToMessage.sender_username : "",
          reply_to_content: this.replyToMessage ? this.replyToMessage.content : "",
        };

        this.messages.push(newMessage);
        // очищаем форму 
        this.messageText = "";
        this.selectedFile = null;
        this.replyToMessage = null;
        this.isInteracting = false;

        this.$nextTick(() => this.scrollToBottom());
      } catch (error) {
        console.error("Error sending message:", error);
      }
    },

    getImageUrl(imagePath) {
      const baseURL = axios.defaults.baseURL;
      return imagePath && imagePath.startsWith("/uploads")
        ? baseURL + imagePath
        : "/default-profile.png";
    },

    async deleteMessage(messageId) {
      if (!confirm("Delete this message?")) return;

      const conversationID = this.$route.params.c_id;
      const token = localStorage.getItem("authToken");

      try {
        await axios.delete(`/conversations/${conversationID}/messages/${messageId}`, {
          headers: { Authorization: `Bearer ${token}` },
        });
        this.messages = this.messages.filter((m) => m.id !== messageId);
      } catch (error) {
        console.error("Error deleting message:", error);
      }
    },

    async deleteComment(message, comment) {
      const token = localStorage.getItem("authToken");
      const conversationID = this.$route.params.c_id;
      if (!token || !conversationID) return;

      try {
        await axios.delete(
          `/conversations/${conversationID}/messages/${message.id}/comments/${comment.id}`,
          { headers: { Authorization: `Bearer ${token}` } }
        );
        // Обновляем список удаляя список локально 
        message.comments = message.comments.filter((c) => c.id !== comment.id);
      } catch (error) {
        console.error("Error deleting comment:", error);
      }
    },
    // показать или скрыть коментарии 
    async toggleComments(message) {
      message.showComments = !message.showComments;

      if (message.showComments) {
        this.isInteracting = true;
        const token = localStorage.getItem("authToken");

        try {
          const response = await axios.get(`/messages/${message.id}/comments`, {
            headers: { Authorization: `Bearer ${token}` },
          });
          message.comments = response.data;
        } catch (error) {
          console.error("Error fetching comments:", error);
        }
      } else {
        this.$nextTick(() => {
          const anyOpen = this.messages.some((m) => m.showComments);
          if (!anyOpen) this.isInteracting = false;
        });
      }
    },

async addEmojiComment(message, emoji) {
  const token = localStorage.getItem("authToken");
  const conversationID = this.$route.params.c_id;
  if (!token || !conversationID) return;

  try {
    await axios.post(
      `/conversations/${conversationID}/messages/${message.id}/comments`,
      { content_type: "emoji", content: emoji },
      { headers: { Authorization: `Bearer ${token}` } }
    );

    const response = await axios.get(`/messages/${message.id}/comments`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    message.comments = response.data;
  } catch (error) {
    console.error("Error adding comment:", error);
  }
}
,

    // панель что показать скрыть перессылкку
    toggleForwardPanel(message) {
      message.showForwardPanel = !message.showForwardPanel;
      this.isInteracting = message.showForwardPanel;
    },
    // Специальная обработка "me" — пересылка себе в Saved Messages 
    async forwardMessageHandler(message) {
      const token = localStorage.getItem("authToken");
      const conversationID = this.$route.params.c_id;
      const currentUserId = localStorage.getItem("userID");

      let targetConversationID;


      if (!token || !conversationID) return;

     if (
  message.forwardTarget.trim().toLowerCase() === "me" ||
  message.forwardTarget.trim() === currentUserId
) {
  try {
    const resp = await axios.get("/users/me/saved", {
      headers: { Authorization: `Bearer ${token}` }
    });
    targetConversationID = resp.data.c_id;
  } catch (error) {
    console.error("Error getting saved messages:", error);
    alert("Could not open Saved Messages Argo");
    return;
  }
}

 else if (!isNaN(message.forwardTarget)) {
        targetConversationID = message.forwardTarget;
      } else {
        targetConversationID = "new";
      }

      try {
        let payload = {};
        if (targetConversationID === "new") {
          payload = { target_username: message.forwardTarget.trim() };
        }

        await axios.post(
          `/conversations/${conversationID}/messages/${message.id}/forward/${targetConversationID}`,
          payload,
          { headers: { Authorization: `Bearer ${token}` } }
        );

        message.status = "forwarded";
        message.showForwardPanel = false;
        message.forwardTarget = "";
        alert("Message forwarded!");
        this.isInteracting = false;
      } catch (error) {
        console.error("Error forwarding:", error);
        alert("Error forwarding message");
      }
    },

    async leaveGroup() {
      const token = localStorage.getItem("authToken");
      const conversationID = this.$route.params.c_id;
      if (!token || !conversationID) return;

      try {
        await axios.delete(`/groups/${conversationID}/leave`, {
          headers: { Authorization: `Bearer ${token}` },
        });
        this.$router.push("/conversations");
      } catch (error) {
        console.error("Error leaving group:", error);
        alert("Error leaving group");
      }
    },

    cancelAddUser() {
      this.showAddUserForm = false;
      this.usernameToAdd = "";
      this.isInteracting = false;
    },

    async addUserToGroup() {
      const token = localStorage.getItem("authToken");
      const conversationID = this.$route.params.c_id;
      if (!token || !conversationID) return;

      if (!this.usernameToAdd.trim()) {
        alert("Please enter a username");
        return;
      }

      try {
        const data = { usernames: [this.usernameToAdd.trim()] };
        await axios.post(`/groups/${conversationID}/members`, data, {
          headers: { Authorization: `Bearer ${token}` },
        });

        alert(`User '${this.usernameToAdd.trim()}' added!`);
        this.usernameToAdd = "";
        this.showAddUserForm = false;
        this.isInteracting = false;
      } catch (error) {
        console.error("Error adding user:", error);
        alert(error.response?.data?.error || "Error adding user");
      }
    },

    cancelSetName() {
      this.showSetNameForm = false;
      this.newGroupName = "";
      this.isInteracting = false;
    },

    async updateGroupName() {
      const token = localStorage.getItem("authToken");
      const conversationID = this.$route.params.c_id;
      if (!token || !conversationID) return;

      if (!this.newGroupName.trim()) {
        alert("Please enter a group name");
        return;
      }

      try {
        await axios.put(
          `/groups/${conversationID}/name`,
          { new_name: this.newGroupName.trim() },
          { headers: { Authorization: `Bearer ${token}` } }
        );

        this.conversationName = this.newGroupName.trim();
        alert("Group name updated!");
        this.newGroupName = "";
        this.showSetNameForm = false;
        this.isInteracting = false;
      } catch (error) {
        console.error("Error updating name:", error);
        alert("Error updating group name");
      }
    },

    handleGroupPhotoChange(event) {
      this.newGroupPhotoFile = event.target.files[0] || null;
      this.isInteracting = true;
    },

    cancelSetPhoto() {
      this.showSetPhotoForm = false;
      this.newGroupPhotoFile = null;
      this.isInteracting = false;
    },

    async updateGroupPhoto() {
      const token = localStorage.getItem("authToken");
      const conversationID = this.$route.params.c_id;
      if (!token || !conversationID) return;

      if (!this.newGroupPhotoFile) {
        alert("Please select a photo");
        return;
      }

      try {
        const formData = new FormData();
        formData.append("photo", this.newGroupPhotoFile);

        const response = await axios.put(
          `/conversations/${conversationID}/set-group-photo`,
          formData,
          {
            headers: {
              Authorization: `Bearer ${token}`,
              "Content-Type": "multipart/form-data",
            },
          }
        );

        if (response.data.photo) {
          const baseURL = axios.defaults.baseURL;
          this.conversationPhoto = response.data.photo.startsWith(baseURL)
            ? response.data.photo
            : `${baseURL}${response.data.photo}`;
        }

        alert("Group photo updated!");
        this.showSetPhotoForm = false;
        this.newGroupPhotoFile = null;
        this.isInteracting = false;
      } catch (error) {
        console.error("Error updating photo:", error);
        alert("Error updating group photo");
      }
    },
  },
};
</script>

<style scoped>
:root {
  --wa-bg-primary: #111b21;
  --wa-bg-secondary: #202c33;
  --wa-bg-hover: #2a3942;
  --wa-border: #2a3942;
  --wa-text-primary: #e9edef;
  --wa-text-secondary: #aebac1;
  --wa-text-muted: #667781;
  --wa-accent: #00a884;
  --wa-accent-hover: #06cf9c;
  --wa-danger: #ea4335;
}

/* Transitions */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.3s;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.25s;
}
.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(20px);
  opacity: 0;
}

.page-wrapper {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
  background-color: var(--wa-bg-primary);
  position: relative;
}

/* ===== HEADER ===== */
.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background-color: var(--wa-bg-secondary);
  border-bottom: 1px solid var(--wa-border);
  position: sticky;
  top: 0;
  z-index: 100;
  min-height: 60px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.back-button {
  background: none;
  border: none;
  color: var(--wa-text-primary);
  font-size: 24px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 50%;
  transition: background-color 0.2s;
  line-height: 1;
}
.back-button:hover {
  background-color: var(--wa-bg-hover);
}

.header-photo {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid var(--wa-border);
}

.header-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.header-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 500;
  color: var(--wa-text-primary);
}

.header-subtitle {
  margin: 0;
  font-size: 0.8rem;
  color: var(--wa-text-muted);
}

.header-actions {
  display: flex;
  gap: 4px;
  align-items: center;
}

.header-icon-btn {
  background: none;
  border: none;
  color: var(--wa-text-secondary);
  font-size: 20px;
  cursor: pointer;
  padding: 8px;
  border-radius: 50%;
  transition: background-color 0.2s;
  line-height: 1;
}
.header-icon-btn:hover {
  background-color: var(--wa-bg-hover);
}

/* ===== MODALS ===== */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 300;
}

.modal-form {
  background-color: var(--wa-bg-secondary);
  border-radius: 12px;
  padding: 24px;
  width: 90%;
  max-width: 400px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.6);
}

.modal-form h3 {
  margin: 0 0 20px 0;
  color: var(--wa-text-primary);
  font-size: 1.3rem;
  font-weight: 500;
}

.modal-form input[type="text"],
.modal-form input[type="file"] {
  width: 100%;
  padding: 12px;
  background-color: var(--wa-bg-hover);
  border: 1px solid var(--wa-border);
  border-radius: 8px;
  color: var(--wa-text-primary);
  font-size: 0.95rem;
  margin-bottom: 20px;
  box-sizing: border-box;
}

.modal-buttons {
  display: flex;
  gap: 12px;
}

.primary-btn {
  flex: 1;
  padding: 12px 24px;
  background-color: var(--wa-accent);
  color: var(--wa-bg-primary);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  font-size: 0.95rem;
  transition: background-color 0.2s;
}
.primary-btn:hover:not(:disabled) {
  background-color: var(--wa-accent-hover);
}
.primary-btn:disabled {
  background-color: var(--wa-bg-hover);
  color: var(--wa-text-muted);
  cursor: not-allowed;
}

.secondary-btn {
  flex: 1;
  padding: 12px 24px;
  background-color: var(--wa-bg-hover);
  color: var(--wa-text-secondary);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  font-size: 0.95rem;
  transition: background-color 0.2s;
}
.secondary-btn:hover {
  background-color: #374955;
}

/* ===== MESSAGES ===== */
.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background-color: var(--wa-bg-primary);
}

.loading-message,
.empty-message {
  text-align: center;
  color: var(--wa-text-secondary);
  padding: 40px;
  font-size: 1rem;
}

.messages-container ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.message-item {
  margin: 15px 0;
  display: flex;
  flex-direction: column;
}

.message-info {
  display: flex;
  gap: 12px;
}

.sender-photo {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
}

.message-content {
  background-color: var(--wa-bg-secondary);
  border-radius: 8px;
  padding: 12px;
  max-width: 70%;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.sender-name {
  font-weight: 600;
  color: var(--wa-accent);
  font-size: 0.9rem;
}

.message-time {
  font-size: 0.75rem;
  color: var(--wa-text-muted);
}

.message-status {
  margin-left: auto;
  font-size: 0.85rem;
}

.status-sent,
.status-delivered {
  color: #8696a0; /* grey check(s): sent = one ✓, delivered = two ✓✓ */
}

.status-read {
  color: #53bdeb; /* blue ✓✓ once the recipient has read the message */
}

.forwarded-label {
  font-size: 0.75rem;
  color: var(--wa-text-muted);
  font-style: italic;
}

.message-actions {
  display: flex;
  gap: 8px;
  margin: 8px 0;
  flex-wrap: wrap;
}

.reply-button,
.forward-button,
.comment-button,
.delete-button {
  padding: 6px 12px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 500;
  transition: all 0.2s;
}

.reply-button {
  background-color: #4a148c;
  color: white;
}
.forward-button {
  background-color: #03a9f4;
  color: white;
}
.comment-button {
  background-color: #9c27b0;
  color: white;
}
.delete-button {
  background-color: var(--wa-danger);
  color: white;
}

.reply-snippet {
  margin: 8px 0;
  padding: 8px 12px;
  background-color: var(--wa-bg-hover);
  border-left: 3px solid var(--wa-accent);
  border-radius: 4px;
  font-size: 0.85rem;
}

.message-text {
  margin: 8px 0 0 0;
  font-size: 0.95rem;
  line-height: 1.5;
  word-wrap: break-word;
  color: var(--wa-text-primary);
}

.message-media {
  width: 100%;
  max-width: 300px;
  height: auto;
  object-fit: cover;
  margin-top: 10px;
  border-radius: 8px;
}

/* ===== FORWARD PANEL ===== */
.forward-panel {
  margin-top: 10px;
  padding: 10px;
  background-color: var(--wa-bg-hover);
  border: 1px solid var(--wa-border);
  border-radius: 6px;
  display: flex;
  gap: 8px;
}

.forward-panel input[type="text"] {
  flex: 1;
  padding: 8px;
  border: 1px solid var(--wa-border);
  border-radius: 6px;
  background-color: var(--wa-bg-secondary);
  color: var(--wa-text-primary);
}

/* ===== COMMENTS ===== */
.comments-section {
  margin-top: 12px;
  padding: 12px;
  border-radius: 8px;
  background-color: var(--wa-bg-hover);
}

.single-comment {
  background-color: var(--wa-bg-secondary);
  margin-bottom: 8px;
  padding: 8px 10px;
  border-radius: 6px;
}

.comment-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
  color: var(--wa-text-primary);
}

.comment-time {
  font-size: 0.75rem;
  color: var(--wa-text-muted);
  margin-left: 8px;
}

.comment-text {
  margin: 0;
  font-size: 0.9rem;
  word-wrap: break-word;
  color: var(--wa-text-secondary);
}

.delete-comment-button {
  background: transparent;
  border: none;
  color: var(--wa-danger);
  font-size: 0.8rem;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
}

.add-comment-form {
  display: flex;
  margin-top: 8px;
  gap: 8px;
}
.add-comment-form button {
  background-color: var(--wa-accent);
  color: var(--wa-bg-primary);
  border: none;
  border-radius: 6px;
  padding: 8px 16px;
  cursor: pointer;
  font-weight: 500;
}

/* ===== REPLY BAR ===== */
.replying-bar {
  background: linear-gradient(90deg, var(--wa-accent) 4px, rgba(0, 168, 132, 0.15) 4px);
  border-radius: 8px;
  padding: 12px 16px;
  margin: 0 16px 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.cancel-reply {
  background: none;
  border: none;
  color: var(--wa-danger);
  font-weight: bold;
  cursor: pointer;
  font-size: 20px;
}

/* ===== INPUT BAR ===== */
.message-input-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background-color: var(--wa-bg-secondary);
  border-top: 1px solid var(--wa-border);
  flex-shrink: 0;
  position: relative;
  z-index: 50;
}

.message-textarea {
  flex: 1;
  padding: 10px 14px;
  border-radius: 8px;
  border: 1px solid var(--wa-border);
  resize: none;
  background-color: var(--wa-bg-hover);
  color: var(--wa-text-primary);
  font-size: 0.95rem;
  min-height: 40px;
  max-height: 120px;
  font-family: inherit;
}

.file-upload-label {
  cursor: pointer;
  font-size: 24px;
  padding: 8px;
  border-radius: 50%;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}
.file-upload-label:hover {
  background-color: var(--wa-bg-hover);
}
.file-input {
  display: none;
}

.more-btn {
  background: none;
  border: none;
  color: var(--wa-text-secondary);
  font-size: 22px;
  padding: 8px;
  border-radius: 50%;
  cursor: pointer;
}
.more-btn:hover {
  background-color: var(--wa-bg-hover);
}

.send-button {
  padding: 10px 24px;
  background-color: var(--wa-accent);
  color: var(--wa-bg-primary);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  font-size: 0.95rem;
}
.send-button:disabled {
  background-color: var(--wa-bg-hover);
  color: var(--wa-text-muted);
  cursor: not-allowed;
}

/* ===== BOTTOM ACTIONS ===== */
.bottom-actions {
  position: fixed;
  left: 12px;
  right: 12px;
  bottom: 78px; /* над input bar */
  z-index: 400;
  background-color: var(--wa-bg-secondary);
  border: 1px solid var(--wa-border);
  border-radius: 12px;
  padding: 10px;
  display: flex;
  gap: 8px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.55);
}

.bottom-actions button {
  flex: 1;
  padding: 10px;
  border-radius: 10px;
  border: none;
  background-color: var(--wa-bg-hover);
  color: var(--wa-text-primary);
  cursor: pointer;
  font-size: 0.9rem;
  white-space: nowrap;
}
.bottom-actions button:hover {
  background-color: #374955;
}
.bottom-actions .danger {
  color: var(--wa-danger);
}

.bottom-actions-backdrop {
  position: fixed;
  inset: 0;
  z-index: 350;
  background: transparent; /* можно rgba(0,0,0,0.2) если хочешь затемнение */
}
</style>
