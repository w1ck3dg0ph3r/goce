#!/bin/env node
import puppeteer from 'puppeteer-core'
import { promisify } from 'util'
import { exec as execCB } from 'child_process'

const exec = promisify(execCB)

const config = {
  chromePath: '/usr/bin/chromium',
  appUrl: 'http://127.0.0.1:9000/',
  zoomPercent: 100,
  removeTemporaryFiles: false,
  compileWebpAnimation: true,
}

;(async () => {
  console.log('launching browser')
  const browser = await puppeteer.launch({
    executablePath: config.chromePath,
    headless: true,
  })

  console.log('opening app')
  const page = await browser.newPage()
  const scaleFactor = config.zoomPercent / 100
  await page.setViewport({
    width: 1920 / scaleFactor,
    height: 1080 / scaleFactor,
    deviceScaleFactor: scaleFactor,
  })
  await page.goto(config.appUrl)
  await sleep(3000)

  console.log('taking screenshots')

  // TODO: Add element ids instead of using pixel coordinates to be more
  // reproducible.

  console.log('1')
  await page.mouse.click(232, 358)
  await page.mouse.move(117, 360)
  await sleep(1000)
  await page.screenshot({ path: './1.png' })

  console.log('2')
  await page.mouse.click(115, 59)
  await sleep(250)
  await page.mouse.click(406, 106)
  await sleep(3000)
  await page.mouse.move(235, 435)
  await sleep(1000)
  await page.screenshot({ path: './2.png' })

  console.log('3')
  await page.mouse.click(1687, 105)
  await sleep(1500)
  await page.screenshot({ path: './3.png' })

  let closed = browser.close()

  if (config.compileWebpAnimation) {
    console.log('compiling webp')
    await exec(`ffmpeg -y\
      -framerate 0.5\
      -i '%d.png'\
      -vf 'scale=iw/2:ih/2'\
      -quality 85\
      -loop 0\
      screenshot.webp`)
  }

  if (config.removeTemporaryFiles) {
    console.log('removing temporaries')
    await exec('rm *.png')
  }

  await closed
})()

async function sleep(ms) {
  await new Promise((res) => setTimeout(res, ms))
}
