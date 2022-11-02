<template>
  <Listbox v-model="model">
    <div class="relative mt-1">
      <ListboxButton
          class="relative w-full cursor-default rounded-lg bg-white py-2 pl-3 pr-10 text-left shadow-md focus:outline-none focus-visible:border-indigo-500 focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-opacity-75 focus-visible:ring-offset-2 focus-visible:ring-offset-orange-300 sm:text-sm"
      >
        <span class="block truncate">{{ model }}</span>
        <span
            class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
        >
            <ChevronUpDownIcon
                class="h-5 w-5 text-gray-400"
                aria-hidden="true"
            />
        </span>
      </ListboxButton>

      <transition
          leave-active-class="transition duration-100 ease-in"
          leave-from-class="opacity-100"
          leave-to-class="opacity-0"
      >
        <ListboxOptions
            class="absolute mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm"
        >
          <ListboxOption
              v-slot="{ active, selected }"
              v-for="opt in options"
              :key="'opt-'+opt"
              :value="opt"
              as="template"
          >
            <li
                :class="[
                  active ? 'bg-indigo-100 text-indigo-900' : 'text-gray-900',
                  'relative cursor-default select-none py-2 pl-10 pr-4',
                ]"
            >
                <span
                    :class="[
                    selected ? 'font-medium' : 'font-normal',
                    'block truncate',
                  ]"
                >{{ opt }}</span>
              <span
                  v-if="selected"
                  class="absolute inset-y-0 left-0 flex items-center pl-3 text-indigo-600"
              >
                  <CheckIcon class="h-5 w-5" aria-hidden="true"/>
              </span>
            </li>
          </ListboxOption>
        </ListboxOptions>
      </transition>
    </div>
  </Listbox>
</template>

<script setup lang="ts">
import {CheckIcon, ChevronUpDownIcon} from "@heroicons/vue/20/solid"

const props = defineProps<{
  modelValue: any
  options: any[]
}>()

const emits = defineEmits(["update:modelValue", "change"])

const model = useVModel(props, "modelValue", emits)

watch(model, value => {
  emits('change', value)
})
</script>
