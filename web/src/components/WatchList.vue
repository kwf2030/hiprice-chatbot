<template>
  <transition enter-active-class="animated slideInLeft">
    <v-layout column>

      <v-list two-line>
        <template v-for="(item, index) in products">
          <v-list-tile avatar ripple :key="item.id">

            <v-list-tile-content>
              <v-list-tile-title>{{ item.title }}</v-list-tile-title>

              <v-layout row>
                <v-flex>
                  <v-list-tile-sub-title>现&emsp;价：</v-list-tile-sub-title>
                </v-flex>
                <v-flex>
                  <v-list-tile-sub-title v-if="item.isPriceRange" class="red--text"> {{ item.currencySymbol }}{{ item.price_low }} - {{ item.currencySymbol }}{{ item.price_high }}</v-list-tile-sub-title>
                  <v-list-tile-sub-title v-else  class="red--text">{{ item.currencySymbol }}{{ item.price }}</v-list-tile-sub-title>
                </v-flex>
              </v-layout>

              <v-layout row>
                <v-flex>
                  <v-list-tile-sub-title>关注价：</v-list-tile-sub-title>
                </v-flex>
                <v-flex>
                  <v-list-tile-sub-title v-if="item.isPriceRange">{{ item.currencySymbol }}{{ item.watch_price_low }} - {{ item.currencySymbol }}{{ item.watch_price_high }}</v-list-tile-sub-title>
                  <v-list-tile-sub-title v-else>{{ item.currencySymbol }}{{ item.watch_price }}</v-list-tile-sub-title>
                </v-flex>
              </v-layout>
            </v-list-tile-content>

            <v-list-tile-action>
              <v-menu bottom left @click.stop.prevent>
                <v-btn icon slot="activator" light>
                  <v-icon>more_vert</v-icon>
                </v-btn>
                <v-list>
                  <v-list-tile>
                    <v-list-tile-title @click="view(item)">
                      <v-list-tile-title>宝贝详情</v-list-tile-title>
                    </v-list-tile-title>
                  </v-list-tile>
                  <v-list-tile @click="openRemind(item)">
                    <v-list-tile-title>提醒设置</v-list-tile-title>
                  </v-list-tile>
                  <v-list-tile @click="unwatch(item)">
                    <v-list-tile-title>不再关注</v-list-tile-title>
                  </v-list-tile>
                </v-list>
              </v-menu>
            </v-list-tile-action>

          </v-list-tile>

          <v-divider v-if="index + 1 < products.length" :key="index"></v-divider>
        </template>
      </v-list>

      <v-subheader v-if="empty">还没有关注任何宝贝哦！</v-subheader>

      <v-layout>
        <v-snackbar v-model="showSnackbar" :timeout="2000" color="error">{{ snackbarText }}</v-snackbar>
      </v-layout>

    </v-layout>
  </transition>
</template>

<script>
  import Bus from '../bus'

  export default {
    data: () => ({
      empty: false,
      products: global.Store.products,
      showSnackbar: false,
      snackbarText: '请求失败，请稍后再试'
    }),

    created () {
      Bus.$emit('tabs', true, '0')
      if (global.Store.user === '') {
        let u = this.getUser()
        if (u === undefined || u === '' || u === 'undefined') {
          console.log('missing user param')
          return
        }
        global.Store.user = u
      }
      if (this.products.length <= 0) {
        this.getWatchList()
      }
    },

    methods: {
      getWatchList () {
        this.$axios.get('watchlist?u=' + global.Store.user).then(resp => {
          if (resp.data.ret !== 0) {
            console.log(resp.data.status)
            this.empty = true
            return
          }
          let arr = resp.data.data.products
          if (arr == null) {
            console.log('getWatchList returns null')
            this.empty = true
            return
          }
          for (let i = 0; i < arr.length; i++) {
            arr[i].isPriceRange = (arr[i].price === -1)
            arr[i].currencySymbol = this.getCurrencySymbol(arr[i].currency)
          }
          this.products = arr
          this.empty = false
          global.Store.products = arr
        }).catch(() => {
          console.log('request failed: getWatchList')
          this.empty = false
          this.showSnackbar = true
        })
      },

      view (product) {
        window.location.href = product.url
      },

      unwatch (product) {
        let data = JSON.stringify({product_id: product.id})
        this.$axios.post('unwatch?u=' + global.Store.user, data).then(resp => {
          if (resp.data.ret !== 0) {
            console.log(resp.data.status)
            return
          }
          let l = this.products.length
          for (let i = 0; i < l; i++) {
            if (this.products[i].id === product.id) {
              this.products.splice(i, 1)
              break
            }
          }
        }).catch(() => {
          console.log('request failed: unwatch for ' + product.id)
          this.showSnackbar = true
        })
      },

      openRemind (product) {
        global.Store.product = product
        this.$router.push({name: 'Remind'})
      },

      getCurrencySymbol (currency) {
        // 0:RMB, 1:JPY, 2:USD, 3:GBP, 4:EUR
        switch (currency) {
          case 0:
            return '￥'
          case 1:
            return '¥'
          case 2:
            return '$'
          case 3:
            return '£'
          case 4:
            return '€'
        }
      },

      getUser () {
        let query = document.location.search.slice(1).split('&')
        for (let i = 0; i < query.length; i++) {
          let kv = query[i].split('=')
          if (kv.length === 2 && kv[0] === 'u') {
            return kv[1]
          }
        }
      }
    }
  }
</script>

<style>
  .animated {
    animation-duration: .2s;
    animation-fill-mode: both;
  }
</style>
