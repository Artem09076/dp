import { theme } from 'antd';

const rostelecomColors = {
  purple: {
    main: '#6B2EE8',
    light: '#8B5CF6',
    dark: '#4C1D95',
  },
  orange: {
    main: '#FF6B00',
    light: '#FF8C42',
    dark: '#E85D00',
  },
  background: {
    white: '#FFFFFF',
    lightGray: '#F5F5F7',
    gray: '#E8E8EC',
  },
  text: {
    primary: '#1A1A2E',
    secondary: '#6B7280',
  },
};

export const antdTheme = {
  token: {
    colorPrimary: rostelecomColors.purple.main,
    colorPrimaryHover: rostelecomColors.purple.light,
    colorPrimaryActive: rostelecomColors.purple.dark,
    colorSuccess: '#52c41a',
    colorWarning: '#faad14',
    colorError: '#ff4d4f',
    colorInfo: rostelecomColors.purple.main,
    colorTextBase: rostelecomColors.text.primary,
    colorBgBase: rostelecomColors.background.white,
    colorBgContainer: rostelecomColors.background.white,
    borderRadius: 8,
    borderRadiusLG: 12,
    borderRadiusSM: 6,
    fontFamily: '"Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif',
    fontSize: 14,
    fontSizeLG: 16,
    fontSizeXL: 20,
    fontSizeHeading1: 38,
    fontSizeHeading2: 30,
    fontSizeHeading3: 24,
    fontSizeHeading4: 20,
    fontSizeHeading5: 16,
    controlHeight: 40,
    controlHeightLG: 48,
    controlHeightSM: 32,
  },
  components: {
    Button: {
      colorPrimary: rostelecomColors.purple.main,
      colorPrimaryHover: rostelecomColors.purple.light,
      colorPrimaryActive: rostelecomColors.purple.dark,
      borderRadius: 8,
      controlHeight: 40,
      controlHeightLG: 48,
      fontWeight: 500,
    },
    Card: {
      borderRadiusLG: 16,
      boxShadow: '0 4px 12px rgba(0, 0, 0, 0.05)',
    },
    Input: {
      borderRadius: 8,
      controlHeight: 40,
      controlHeightLG: 48,
    },
    Select: {
      borderRadius: 8,
      controlHeight: 40,
      controlHeightLG: 48,
    },
    Layout: {
      headerBg: rostelecomColors.purple.main,
      headerColor: '#FFFFFF',
      bodyBg: rostelecomColors.background.lightGray,
    },
    Menu: {
      itemBg: 'transparent',
      itemColor: '#FFFFFF',
      itemHoverBg: 'rgba(255, 255, 255, 0.1)',
      itemSelectedBg: 'rgba(255, 255, 255, 0.15)',
      itemSelectedColor: '#FFFFFF',
    },
  },
};