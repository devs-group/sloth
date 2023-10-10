<template>
    <div class="h-[calc(100vh-3rem)]">
        <Notification class="absolute bottom-6 right-6 w-96 z-20"></Notification>
        <TopBar></TopBar>
        <div class="flex flex-row h-full">
            <Navigation></Navigation>
            <slot></slot>
        </div>
        
    </div>
</template>