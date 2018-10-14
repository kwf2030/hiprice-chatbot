<template>
  <v-layout column>

    <v-list two-line>
      <template v-for="(bot, index) in bots">
        <v-list-tile avatar ripple :key="bot.uin">

          <v-list-tile-content>
            <v-list-tile-title>{{ bot.nickname }}</v-list-tile-title>

            <v-layout row>
              <v-flex>
                <v-list-tile-sub-title>UIN：</v-list-tile-sub-title>
              </v-flex>
              <v-flex>
                <v-list-tile-sub-title> {{ bot.uin }}</v-list-tile-sub-title>
              </v-flex>
            </v-layout>

            <v-layout row>
              <v-flex>
                <v-list-tile-sub-title>登录时间：</v-list-tile-sub-title>
              </v-flex>
              <v-flex>
                <v-list-tile-sub-title>{{ bot.start_time }}</v-list-tile-sub-title>
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
                  <v-list-tile-title @click="showLogoutDialog(bot)">
                    <v-list-tile-title>退出</v-list-tile-title>
                  </v-list-tile-title>
                </v-list-tile>
              </v-list>
            </v-menu>
          </v-list-tile-action>

        </v-list-tile>
        <v-divider v-if="index + 1 < bots.length" :key="index"></v-divider>
      </template>
    </v-list>

    <v-layout row justify-center>
      <v-dialog v-model="qrCodeDialog">
        <v-card>
          <v-card-title>
            <div class="headline">
              <v-flex xs12 class="text-xs-center">
                <v-card-text v-if="qrCodeState < 0" class="red--text"><h6>二维码加载失败</h6></v-card-text>
                <v-card-text v-else-if="qrCodeState === 0">等待扫描...</v-card-text>
                <v-card-text v-else-if="qrCodeState === 1">等待确认...</v-card-text>
                <v-card-text v-else-if="qrCodeState === 2">正在登录...</v-card-text>
              </v-flex>
            </div>
            <v-spacer></v-spacer>
            <v-btn icon @click.native="qrCodeDialog = false">
              <v-icon>close</v-icon>
            </v-btn>
          </v-card-title>

          <v-flex v-if="qrCodeLoading" class="pb-3" d-flex justify-center>
            <v-progress-circular indeterminate></v-progress-circular>
          </v-flex>

          <v-flex v-if="qrCodeUrl !== ''" class="pb-3" d-flex justify-center>
            <v-img :src="qrCodeUrl" height="320"></v-img>
          </v-flex>
        </v-card>
      </v-dialog>
    </v-layout>

    <v-layout row justify-center>
      <v-dialog v-model="logoutDialog">
        <v-card>
          <v-card-title class="headline">确定退出？</v-card-title>
          <v-card-actions>
            <v-spacer></v-spacer>
            <v-btn flat="flat" @click.native="dismissLogoutDialog()">取消</v-btn>
            <v-btn class="red--text" flat="flat" @click.native="logout()">退出</v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </v-layout>

  </v-layout>
</template>

<script>
  import Bus from '../bus'

  export default {
    data: () => ({
      bots: [],
      bot: null,
      qrCodeUrl: '',
      qrCodeDialog: false,
      qrCodeLoading: true,
      qrCodeState: 0,
      logoutDialog: false,
      logoutUin: 0
    }),

    created () {
      Bus.$on('fab.click', () => {
        this.login()
      })
      this.getBots()
    },

    methods: {
      getBots () {
        this.$axios.get('/admin/api/bots').then(resp => {
          if (resp.data.ret !== 0) {
            console.log(resp.data.status)
            return
          }
          this.bots = resp.data.data.bots
        }).catch(() => {
          console.log('request failed: getBots')
        })
      },

      login () {
        this.qrCodeDialog = true
        this.qrCodeLoading = true
        this.$axios.post('/admin/api/bot').then(resp => {
          if (resp.data.ret !== 0) {
            console.log(resp.data.status)
            return
          }
          this.qrCodeUrl = resp.data.data.qrcode
          this.qrCodeLoading = false
          let token = setInterval(() => {
            if (this.qrCodeState < 0) {
              clearInterval(token)
              this.clearLogin()
              return
            }
            if (this.bot !== null) {
              clearInterval(token)
              this.bots.push(this.bot)
              this.clearLogin()
              return
            }
            this.getLoginState()
          }, 1000)
        }).catch(() => {
          console.log('request failed: login')
          this.qrCodeLoading = false
          this.qrCodeState = -1
        })
      },

      getLoginState () {
        let uuid = this.qrCodeUrl.slice(this.qrCodeUrl.lastIndexOf('/') + 1)
        this.$axios.get('/admin/api/bot?uuid=' + uuid).then(resp => {
          if (resp.data.ret !== 0) {
            console.log(resp.data.status)
            return
          }
          this.qrCodeState = resp.data.data.state
          let b = resp.data.data.bot
          if (b !== undefined && b !== null) {
            this.bot = b
          }
        }).catch(() => {
          console.log('request failed: getLoginState')
        })
      },

      clearLogin () {
        this.bot = null
        this.qrCodeUrl = ''
        this.qrCodeDialog = false
        this.qrCodeLoading = true
        this.qrCodeState = 0
      },

      showLogoutDialog (bot) {
        this.logoutUin = bot.uin
        this.logoutDialog = true
      },

      dismissLogoutDialog () {
        this.logoutUin = 0
        this.logoutDialog = false
      },

      logout () {
        this.$axios.delete('/admin/api/bot?uin=' + this.logoutUin).then(resp => {
          if (resp.data.ret !== 0) {
            console.log(resp.data.status)
            return
          }
          for (let i = 0; i < this.bots.length; i++) {
            if (this.bots[i].uin === this.logoutUin) {
              this.bots.splice(i, 1)
              break
            }
          }
          this.dismissLogoutDialog()
        }).catch(() => {
          console.log('request failed: logout')
        })
      }
    }
  }
</script>
