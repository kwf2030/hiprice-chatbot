<template>
  <transition enter-active-class="animated slideInUp">
    <v-layout column>

      <v-layout row d-flex>
        <v-card flat>
          <v-toolbar dark dense color="blue">
            <v-btn icon dark @click.native="close()">
              <v-icon>close</v-icon>
            </v-btn>
            <v-toolbar-title>提醒设置</v-toolbar-title>
            <v-spacer></v-spacer>
            <v-toolbar-items>
              <v-btn dark flat @click="save()">完成</v-btn>
            </v-toolbar-items>
          </v-toolbar>

          <v-layout v-if="product !== null" column wrap ma-3>
            <v-flex xs12 class="title">
              <span>{{ product.title }}</span>
            </v-flex>

            <v-layout row wrap mt-3>
              <v-flex xs6 class="text-xs-right grey--text">
                <span>现&emsp;&emsp;&emsp;价：</span>
              </v-flex>
              <v-flex xs6 class="text-xs-left grey--text">
                <span v-if="product.isPriceRange" class="red--text">{{ product.currencySymbol }}{{ product.price_low }} - {{ product.currencySymbol }}{{ product.price_high }}</span>
                <span v-else class="red--text">{{ product.currencySymbol }}{{ product.price }}</span>
              </v-flex>
            </v-layout>

            <v-layout row wrap>
              <v-flex xs6 class="text-xs-right grey--text">
                <span>关&emsp;注&emsp;价：</span>
              </v-flex>
              <v-flex xs6 class="text-xs-left grey--text">
                <span v-if="product.isPriceRange">{{ product.currencySymbol }}{{ product.watch_price_low }} - {{ product.currencySymbol }}{{ product.watch_price_high }}</span>
                <span v-else>{{ product.currencySymbol }}{{ product.watch_price }}</span>
              </v-flex>
            </v-layout>

            <v-layout row wrap>
              <v-flex xs6 class="text-xs-right grey--text">
                <span>历史最高价：</span>
              </v-flex>
              <v-flex xs6 class="text-xs-left grey--text">
                <span>暂不可用</span>
                <!--<span v-if="product.isPriceRange">{{ product.currencySymbol }}{{ product.highest_price_low }} - {{ product.currencySymbol }}{{ product.highest_price_high }}</span>
                <span v-else>{{ product.currencySymbol }}{{ product.highest_price }}</span>-->
              </v-flex>
            </v-layout>

            <v-layout row wrap>
              <v-flex xs6 class="text-xs-right grey--text">
                <span>历史最低价：</span>
              </v-flex>
              <v-flex xs6 class="text-xs-left grey--text">
                <span>暂不可用</span>
                <!--<span v-if="product.isPriceRange">{{ product.currencySymbol }}{{ product.lowest_price_low }} - {{ product.currencySymbol }}{{ product.lowest_price_high }}</span>
                <span v-else>{{ product.currencySymbol }}{{ product.lowest_price }}</span>-->
              </v-flex>
            </v-layout>
          </v-layout>

          <v-layout v-if="!product.isPriceRange" row wrap>
            <v-flex xs5 offset-xs1 mt-4>
              <span class="subheading grey--text">降价提醒</span>
            </v-flex>
            <v-flex xs5>
              <v-select v-model="remind.decreaseOption" :items="remind.options" label="提醒方式" return-object single-line clearable></v-select>
            </v-flex>
            <v-flex xs5 offset-xs6 v-if="remind.decreaseOption !== null && remind.decreaseOption.value === 1">
              <v-text-field v-model="remind.decreasePrice" prefix="￥" label="价格低于" type="number" :rules="[remind.priceValidate]"></v-text-field>
            </v-flex>
            <v-flex xs5 offset-xs6 v-if="remind.decreaseOption !== null && remind.decreaseOption.value === 2">
              <v-text-field v-model="remind.decreasePrice" suffix="%" label="降幅超过" type="number" :rules="[remind.rateValidate]"></v-text-field>
            </v-flex>
          </v-layout>

          <v-layout v-if="!product.isPriceRange" row wrap>
            <v-flex xs5 offset-xs1 mt-4>
              <span class="subheading grey--text">涨价提醒</span>
            </v-flex>
            <v-flex xs5>
              <v-select v-model="remind.increaseOption" :items="remind.options" label="提醒方式" return-object single-line clearable></v-select>
            </v-flex>
            <v-flex xs5 offset-xs6 v-if="remind.increaseOption !== null && remind.increaseOption.value === 1">
              <v-text-field v-model="remind.increasePrice" prefix="￥" label="价格高于" type="number" :rules="[remind.priceValidate]"></v-text-field>
            </v-flex>
            <v-flex xs5 offset-xs6 v-if="remind.increaseOption !== null && remind.increaseOption.value === 2">
              <v-text-field v-model="remind.increasePrice" suffix="%" label="涨幅超过" type="number" :rules="[remind.rateValidate]"></v-text-field>
            </v-flex>
          </v-layout>

        </v-card>
      </v-layout>

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
      product: global.Store.product,
      remind: {
        options: [{value: 1, text: '按价格'}, {value: 2, text: '按比例'}],
        decreaseOption: null,
        decreasePrice: 0,
        increaseOption: null,
        increasePrice: 0,
        priceValidate: value => {
          let v = parseFloat(value)
          if (isNaN(v)) {
            return true
          }
          if (Math.floor(v) !== v) {
            return '必须是整数'
          }
          if (v < 1 || v > 999999) {
            return '必须在1-999999之间'
          }
          return true
        },
        rateValidate: value => {
          let v = parseFloat(value)
          if (isNaN(v)) {
            return true
          }
          if (Math.floor(v) !== v) {
            return '必须是整数'
          }
          if (v < 1 || v > 99) {
            return '必须在1-99之间'
          }
          return true
        }
      },
      showSnackbar: false,
      snackbarText: '请求失败，请稍后再试'
    }),

    created () {
      Bus.$emit('tabs', false)
      if (global.Store.user === '') {
        console.log('missing user param')
        return
      }
      if (!this.product.isPriceRange) {
        this.getRemind(this.product.id)
      }
    },

    methods: {
      getRemind (productID) {
        this.$axios.get('remind?u=' + global.Store.user + '&p=' + productID).then(resp => {
          if (resp.data.ret !== 0) {
            console.log(resp.data.status)
            return
          }
          let r = resp.data.data
          switch (r.remind_decrease_option) {
            case 1:
              this.remind.decreaseOption = this.remind.options[0]
              break
            case 2:
              this.remind.decreaseOption = this.remind.options[1]
              break
          }
          this.remind.decreasePrice = r.remind_decrease_price
          switch (r.remind_increase_option) {
            case 1:
              this.remind.increaseOption = this.remind.options[0]
              break
            case 2:
              this.remind.increaseOption = this.remind.options[1]
              break
          }
          this.remind.increasePrice = r.remind_increase_price
        }).catch(() => {
          console.log('request failed: getRemind')
          this.showSnackbar = true
        })
      },

      close () {
        global.Store.product = null
        this.$router.back()
      },

      save () {
        if (this.product.isPriceRange) {
          this.close()
          return
        }
        let dov = 0
        let iov = 0
        if (this.remind.decreaseOption !== null) {
          dov = this.remind.decreaseOption.value
        }
        if (this.remind.increaseOption !== null) {
          iov = this.remind.increaseOption.value
        }
        if (dov !== 0) {
          if (this.remind.decreasePrice <= 0) {
            return
          }
          if (dov === 1 && this.remind.decreasePrice > 999999) {
            return
          }
          if (dov === 2 && this.remind.decreasePrice > 99) {
            return
          }
        }
        if (iov !== 0) {
          if (this.remind.increasePrice <= 0) {
            return
          }
          if (iov === 1 && this.remind.increasePrice > 999999) {
            return
          }
          if (iov === 2 && this.remind.increasePrice > 99) {
            return
          }
        }
        let data = JSON.stringify({
          product_id: this.product.id,
          remind_decrease_option: this.remind.decreaseOption === null ? 0 : this.remind.decreaseOption.value,
          remind_decrease_price: parseFloat(this.remind.decreasePrice),
          remind_increase_option: this.remind.increaseOption === null ? 0 : this.remind.increaseOption.value,
          remind_increase_price: parseFloat(this.remind.increasePrice)
        })
        this.$axios.post('remind?u=' + global.Store.user, data).then(resp => {
          if (resp.data.ret !== 0) {
            console.log(resp.data.status)
            return
          }
          this.close()
        }).catch(() => {
          console.log('request failed: save for ' + this.product.id)
          this.showSnackbar = true
        })
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
