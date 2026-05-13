<script setup>
import { RouterLink, RouterView, useRoute } from 'vue-router'
import SidebarContent from "./components/SidebarContent.vue";
const route = useRoute()
</script>
<script>
export default {}
</script>

<template>
  <header v-if="route.path !== '/'" class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
    <RouterLink class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" to="/home">
      WasaText
    </RouterLink>
    <!--<a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" href="#/home">WasaText</a>
		<button class="navbar-toggler position-absolute d-md-none collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#sidebarMenu" aria-controls="sidebarMenu" aria-expanded="false" aria-label="Toggle navigation">
			<span class="navbar-toggler-icon"></span>
		</button>-->
  </header>


  <div class="container-fluid">
    <div class="row">
      <nav v-if="route.path.includes('conversation') || route.path.includes('group')" id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
        <SidebarContent />
      </nav>

      <main :class="['px-md-4', (route.path === '/' || route.path === '/home') ? 'col-12' : 'col-md-9 ms-sm-auto col-lg-10']">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<style>
.sidebar {
	height: 100vh;
	display: flex;
	flex-direction: column;
}
</style>
