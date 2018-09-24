<template>
  <v-app>

    <v-toolbar color="blue" dark fixed app>
      <v-toolbar-side-icon></v-toolbar-side-icon>
      <v-toolbar-title>HiPrice</v-toolbar-title>
    </v-toolbar>

    <v-content>
      <router-view/>
    </v-content>

    <v-fab-transition>
      <v-btn v-show="showFab" dark fab bottom right fixed @click.native.stop="onFabClick()" class="red">
        <v-icon>add</v-icon>
      </v-btn>
    </v-fab-transition>

  </v-app>
</template>

<script>
  import Bus from './bus'

  export default {
    name: 'App',

    data: () => ({
      showFab: true
    }),

    created () {
      Bus.$on('fab', show => {
        if (show !== undefined && show !== this.showFab) {
          this.showFab = show
        }
      })
    },

    methods: {
      onFabClick () {
        Bus.$emit('fab-click')
      }
    }
  }
</script>
