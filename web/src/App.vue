<template>
  <v-app>
    <v-tabs v-show="showTabs" v-model="tabIndex" light fixed-tabs slider-color="blue">
      <v-tab @click="onTabClick('WatchList')">我的关注</v-tab>
      <v-tab @click="onTabClick('Settings')">设置</v-tab>
    </v-tabs>

    <v-content>
      <router-view/>
    </v-content>
  </v-app>
</template>

<script>
  import Bus from './bus'

  export default {
    name: 'App',

    data: () => ({
      showTabs: true,
      tabIndex: '0'
    }),

    created () {
      Bus.$on('tabs', (show, index) => {
        if (show !== undefined && show !== this.showTabs) {
          this.showTabs = show
        }
        if (index !== undefined && index !== this.tabIndex) {
          this.tabIndex = index
        }
      })
    },

    methods: {
      onTabClick (route) {
        if (this.$route.name === route) {
          return
        }
        switch (route) {
          case 'WatchList':
            this.$router.back()
            break
          case 'Settings':
            this.$router.push({name: 'Settings'})
            break
        }
      }
    }
  }
</script>
