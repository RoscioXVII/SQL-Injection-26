<script setup>
/*
Realized by @luckignolo32 (GitHub) - MIT License
*/
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { doRegister } from "../services/axios";

const router = useRouter();
const username = ref('');
const password = ref('');
const errorMessage = ref(null);

/*
register submits the registration form to the backend. On success it stores the
returned session credentials in sessionStorage and forwards the user to /home,
realising the requested auto-login behaviour. On failure it surfaces the backend
error message inside the local errorMessage ref so the template can render it.
*/
const register = async () => {
    errorMessage.value = null;
    if(username.value === '' || password.value === ''){
        errorMessage.value = "Inserire username e password";
        return;
    }
    try{
        const userData = await doRegister(username.value, password.value);

        sessionStorage.setItem('username', username.value);
        sessionStorage.setItem('userId', userData.userId);
        sessionStorage.setItem('token', userData.token);

        router.push('/home');
    } catch(err){
        errorMessage.value = err.message;
    }
};
</script>

<template>
  <div class="d-flex justify-content-center align-items-center vh-100">
    <div class="card p-4 shadow-sm" style="max-width: 400px; width: 100%;">
      <h2 class="text-center mb-4">Registrazione</h2>
      <div class="mb-3">
        <label for="username" class="form-label">Username</label>
        <input
          id="username"
          v-model="username"
          type="text"
          class="form-control"
          placeholder="Scegli uno username"
          @keyup.enter="register"
        >
      </div>

      <div class="mb-3">
       <label for="password" class="form-label">Password</label>
        <input
         id="password"
         v-model="password"
         type="password"
         class="form-control"
         placeholder="Scegli una password"
         @keyup.enter="register"
         >
      </div>

      <ErrorMsg v-if="errorMessage" :msg="errorMessage" />

      <button
        type="button"
        class="btn btn-primary w-100"
        :disabled="!username || !password"
        @click="register"
      >
        Registrati
      </button>

      <div class="text-center mt-3">
        <RouterLink to="/">Hai già un account? Accedi</RouterLink>
      </div>
    </div>
  </div>
</template>

<style>
</style>
