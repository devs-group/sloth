<template>
    <div class="h-[calc(100vh-3rem)]">
        <TopBar></TopBar>
        <div class="flex flex-row h-full">
            <Navigation></Navigation>
            <slot></slot>
        </div>
        
    </div>
</template>