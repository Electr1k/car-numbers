<template>
  <div class="app">
    <AppHeader />
    <main>
      <NuxtPage />
    </main>
    <SellerPanel />
  </div>
</template>

<style scoped>
.app { min-height: 100vh; display: flex; flex-direction: column; }
main { flex: 1; }
</style>
