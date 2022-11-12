<template>
  <TransitionRoot appear :show="isOpen" as="template">
    <Dialog as="div" class="relative z-10">
      <TransitionChild as="template" enter="duration-300 ease-out" enter-from="opacity-0" enter-to="opacity-100"
        leave="duration-200 ease-in" leave-from="opacity-100" leave-to="opacity-0">
        <div class="fixed inset-0 bg-black bg-opacity-25" />
      </TransitionChild>

      <div class="fixed inset-0 overflow-y-auto">
        <div class="flex min-h-full items-center justify-center p-4 text-center">
          <TransitionChild as="template" enter="duration-300 ease-out" enter-from="opacity-0 scale-95"
            enter-to="opacity-100 scale-100" leave="duration-200 ease-in" leave-from="opacity-100 scale-100"
            leave-to="opacity-0 scale-95">
            <DialogPanel
              class="w-full max-w-md transform overflow-hidden rounded-2xl bg-white p-6 text-left align-middle shadow-xl transition-all">
              <DialogTitle as="h3" class="text-lg font-medium leading-6 text-gray-900">
                {{ t('please-login') }}
              </DialogTitle>
              <!-- account input -->
              <div class="mt-6 space-y-4 rounded-md shadow-sm">
                <div>
                  <label for="username" class="sr-only">{{ t('username') }}</label>
                  <input v-model="account.username" id="username" name="username" type="text" autocomplete="text"
                    required class="text-input" :placeholder="t('username')" />
                </div>
                <div>
                  <label for="password" class="sr-only">{{ t('password') }}</label>
                  <input v-model="account.password" id="password" name="password" type="password"
                    autocomplete="current-password" required class="text-input" :placeholder="t('password')" />
                </div>
              </div>

              <div class="mt-10">
                <button type="button"
                  class="inline-flex justify-center rounded-md border border-transparent bg-indigo-100 px-4 py-2 text-sm font-medium text-indigo-900 hover:bg-indigo-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500 focus-visible:ring-offset-2"
                  @click="doLogin">
                  {{ t('login') }}
                </button>
              </div>
            </DialogPanel>
          </TransitionChild>
        </div>
      </div>
    </Dialog>
  </TransitionRoot>
</template>
  
<script setup lang="ts">
const props = withDefaults(defineProps<{
  modelValue: boolean
}>(), {
  modelValue: false
})

const account = reactive({
  "username": "",
  "password": ""
})

const emits = defineEmits(['update:modelValue'])

const isOpen = useVModel(props, 'modelValue', emits)

const { t } = useI18n({
  inheritLocale: true
})

function doLogin() {
  useStore().setAuth(account.password, account.password)
  isOpen.value = false
}

</script>

<style scoped>
.text-input {
  @apply block w-full appearance-none rounded-md border border-gray-300 px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:outline-none focus:ring-indigo-500 sm:text-sm
}
</style>

<i18n lang="yaml">
en:
  please-login: 'Please login to authenticate'
  username: 'Username'
  password: 'Password'
  login: 'Sgin in'
zh:
  please-login: '请登录确认身份'
  username: '用户名'
  password: '密码'
  login: '登陆'
</i18n>
  