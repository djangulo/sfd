let user = Cypress.env("BASICAUTH_USER") ? Cypress.env("BASICAUTH_USER") : "";
let pass = Cypress.env("BASICAUTH_PASS") ? Cypress.env("BASICAUTH_PASS") : "";
describe("Item endpoint tests", function () {
  before(() => {
    // runs once before all tests in the block
    cy.clearCookies();
  });
  beforeEach(() => {
    Cypress.Cookies.preserveOnce("sfd-session-id");
  });
  it("can navigate to site", () => {
    cy.visit("/articulos/test-item-120", {
      auth: { username: user, password: pass },
    });
  });
  it(`should have an enabled button to place bid`, () => {
    cy.get("button.ui.primary.button").should("be.enabled");
  });

  it(`clicking the button should redirect to login view`, () => {
    cy.get("button.ui.primary.button").click();
    cy.location("search").should("contain", "next");
  });

  it(`once logged in should redirect back to the item`, () => {
    cy.get(`input[name="username"]`).clear();
    cy.get(`input[name="password"]`).clear();
    cy.get(`input[name="username"]`).type("testuser-00");
    cy.get(`input[name="password"]`).type(`testuser-00`);
    cy.get(`button[type="submit"]`).click();
    cy.location("pathname").should(
      "eq",
      "/articulos/test-item-120/colocar-apuesta"
    );
  });
  it(`should redirect back to the item once bid is placed`, () => {
    // const input = cy.get(`input[name="amount"]`);
    // console.log(input.invoke("attrs", "min").valueOf());
    // input.type(input.invoke("attrs", "min"));
    cy.get(`button[type="submit"]`).click();
  });

  // it(`clicking the empty form returns errors`, () => {
  //   cy.get(`button[type="submit"]`).click();
  //   cy.get(`.ui.error.message`).contains("no puede estar vacío");
  // });
  // it(`enters some bogus data`, () => {
  //   cy.get(`input[name="username"]`).type("madeupusername");
  //   cy.get(`input[name="password"]`).type(`madeuppassowrd`);
  //   cy.get(`button[type="submit"]`).click();
  //   cy.get(`.ui.error.message`).contains(
  //     "nombre de usuario o contraseña inválido"
  //   );
  // });
  // it(`has the session cookie "sfd-session-id"`, () => {
  //   cy.getCookie("sfd-session-id").should("exist");
  // });
  // it(`should be able to logout`, () => {
  //   cy.get("nav").within(() => {
  //     cy.get(`i.dropdown.icon`).click();
  //     cy.get(`a[href="/cerrar-sesion"]`).click();
  //   });
  // });
  // it(`session cookie "sfd-session-id" should not exist`, () => {
  //   cy.getCookie("sfd-session-id").should("not.exist");
  // });
  // it(`should redirect to "/"`, () => {
  //   cy.location("pathname").should("eq", "/");
  // });
});
