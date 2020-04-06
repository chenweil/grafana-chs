import {
  InputPageObject,
  ClickablePageObject,
  Selector,
  InputPageObjectType,
  ClickablePageObjectType,
} from 'e2e-test/core/pageObjects';
import { TestPage } from 'e2e-test/core/pages';

export interface LoginPage {
  username: InputPageObjectType;
  password: InputPageObjectType;
  submit: ClickablePageObjectType;
}

export const loginPage = new TestPage<LoginPage>({
  url: '/login',
  pageObjects: {
    username: new InputPageObject(Selector.fromAriaLabel('用户名输入字段')),
    password: new InputPageObject(Selector.fromAriaLabel('密码输入字段')),
    submit: new ClickablePageObject(Selector.fromAriaLabel('登录按钮')),
  },
});
