<template>
	<Popover class="relative">
		<!-- TODO slot not work-->
		<PopoverButton>
			<slot name="botton"></slot>
		</PopoverButton>
		<Transition enter-active-class="transition duration-200 ease-out" enter-from-class="translate-y-1 opacity-0"
			enter-to-class="translate-y-0 opacity-100" leave-active-class="transition duration-150 ease-in"
			leave-from-class="translate-y-0 opacity-100" leave-to-class="translate-y-1 opacity-0">
			<PopoverPanel
				class="absolute overflow-hidden rounded-lg shadow-lg ring-1 ring-black ring-opacity-5 left-1/2 z-10 mt-3 w-screen max-w-[10rem] -translate-x-full transform px-1 sm:px-0">
				<slot name="panel"></slot>
			</PopoverPanel>
		</Transition>
	</Popover>
</template>