import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore('user', () => {
    const isLoggedIn = ref(false)
    const loggedInUser = ref('')

    // TODO: implement login/logout/status

    return { isLoggedIn, loggedInUser }
})
